package porter

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
)

func (p *Porter) Build() error {
	err := p.Config.LoadManifest("porter.yaml")
	if err != nil {
		return nil
	}

	if err := p.generateDockerFile(); err != nil {
		return fmt.Errorf("unable to generate Dockerfile: %s", err)
	}

	digest, err := p.buildInvocationImage(context.Background())
	if err != nil {
		return fmt.Errorf("unable to build CNAB invocation image: %s", err)
	}

	taggedImage, err := p.rewriteImageWithDigest(p.Config.Manifest.Image, digest)
	if err != nil {
		return fmt.Errorf("unable to regenerate tag: %s", err)
	}

	return p.buildBundle(taggedImage, digest)
}

func (p *Porter) generateDockerFile() error {
	fmt.Printf("\nGenerating Dockerfile =======>\n")
	f, err := p.Config.FileSystem.OpenFile("Dockerfile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open Dockerfile: %s", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	err = p.addDockerBaseImage(w)
	err = p.addCNAB(w)
	err = p.addPorterYAML(w)
	err = p.addMixins(w)
	err = p.addRun(w)

	defer w.Flush()
	return err
}

func (p *Porter) addDockerBaseImage(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "FROM ubuntu:latest\n"); err != nil {
		return fmt.Errorf("couldn't write docker base image: %s", err)
	}
	return nil
}

func (p *Porter) addPorterYAML(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "COPY porter.yaml cnab/app/porter.yaml\n"); err != nil {
		return fmt.Errorf("couldn't write porter.yaml: %s", err)
	}
	return nil
}

func (p *Porter) addCNAB(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "ADD cnab/ cnab/\n"); err != nil {
		return fmt.Errorf("couldn't write docker base image: %s", err)
	}
	return nil
}

func (p *Porter) addMixins(w io.Writer) error {

	// Always copy in porter
	homedir, _ := p.GetHomeDir()
	mixinDir := fmt.Sprintf("%s/mixins", homedir)
	porterPath := fmt.Sprintf("%s/%s", mixinDir, "porter")
	porterMixin, err := p.Config.FileSystem.Open(porterPath)
	if err != nil {
		return fmt.Errorf("couldn't open porter for container build: %s", err)
	}
	defer porterMixin.Close()
	f, err := p.Config.FileSystem.OpenFile("cnab/app/porter", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("couldn't open porter in build path for container build: %s", err)
	}
	defer f.Close()
	_, err = io.Copy(f, porterMixin)
	if err != nil {
		return fmt.Errorf("couldn't write mixin: %s", err)
	}

	cnabMixins := "cnab/app/mixins"
	mixinDirTemplate := "cnab/app/mixins/%s"
	mixinExecTemplate := "cnab/app/mixins/%s/%s"
	mixinsDirExists, err := p.Config.FileSystem.DirExists(cnabMixins)
	if err != nil {
		return fmt.Errorf("couldn't verify mixins directory: %s", err)
	}
	if !mixinsDirExists {
		p.Config.FileSystem.Mkdir(cnabMixins, 0755)
	}
	fmt.Printf("Processing mixins ===> \n")
	for _, mixin := range p.Manifest.Mixins {
		err := func() error {
			fmt.Printf("Processing mixin %s ===> \n", mixin)
			fmt.Printf("Reading: %s\n", fmt.Sprintf("%s/%s", mixinDir, mixin))
			fmt.Printf("Writing: %s\n", fmt.Sprintf(mixinExecTemplate, mixin, mixin))

			mixinsDirExists, err := p.Config.FileSystem.DirExists(fmt.Sprintf(mixinDirTemplate, mixin))
			if err != nil {
				return fmt.Errorf("couldn't verify mixins directory: %s", err)
			}

			if !mixinsDirExists {
				p.Config.FileSystem.Mkdir(fmt.Sprintf(mixinDirTemplate, mixin), 0755)
			}

			mixinExec, err := p.Config.FileSystem.Open(fmt.Sprintf("%s/%s", mixinDir, mixin))
			if err != nil {
				return fmt.Errorf("couldn't open mixin for container build: %s", err)
			}
			defer mixinExec.Close()

			f, err := p.Config.FileSystem.OpenFile(fmt.Sprintf(mixinExecTemplate, mixin, mixin), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
			if err != nil {
				return fmt.Errorf("couldn't open mixin in build path container build: %s", err)
			}
			defer f.Close()

			_, err = io.Copy(f, mixinExec)
			return errors.Wrapf(err, "couldn't write mixin %q", mixin)
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Porter) addRun(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "CMD [\"/cnab/app/run\"]"); err != nil {
		return fmt.Errorf("couldn't write docker base image: %s", err)
	}
	return nil
}

func (p *Porter) buildInvocationImage(ctx context.Context) (string, error) {
	fmt.Printf("\nStarting Invocation Image Build =======> \n")
	path, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error %s", err)
	}
	buildOptions := types.ImageBuildOptions{
		SuppressOutput: false,
		PullParent:     false,
		Tags:           []string{p.Config.Manifest.Image},
		Dockerfile:     "Dockerfile",
	}
	tar, _ := archive.TarWithOptions(path, &archive.TarOptions{})

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", fmt.Errorf("cannot create Docker client: %v", err)
	}
	cli.NegotiateAPIVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("cannot update Docker client: %v", err)
	}

	response, err := cli.ImageBuild(context.Background(), tar, buildOptions)
	if err != nil {
		return "", err
	}
	termFd, _ := term.GetFdInfo(os.Stdout)
	// Setting this to false here because Moby os.Exit(1) all over the place and this fails on WSL (only)
	// when Term is true.
	isTerm := false
	err = jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		return "", err
	}

	authConfig := types.AuthConfig{
		Username: os.Getenv("DOCKER_USER"),
		Password: os.Getenv("DOCKER_PASSWORD"),
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", fmt.Errorf("unable to build Docker auth:%s", err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	pushResponse, err := cli.ImagePush(ctx, p.Config.Manifest.Image, types.ImagePushOptions{
		All:          true,
		RegistryAuth: authStr,
	})
	if err != nil {
		return "", err
	}
	defer pushResponse.Close()
	err = jsonmessage.DisplayJSONMessagesStream(pushResponse, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		return "", err
	}
	dist, err := cli.DistributionInspect(ctx, p.Config.Manifest.Image, "")
	if err != nil {
		return "", fmt.Errorf("unable to inspect image: %s", err)
	}
	return string(dist.Descriptor.Digest), nil
}

func (p *Porter) rewriteImageWithDigest(InvocationImage string, digest string) (string, error) {
	ref, err := reference.Parse(InvocationImage)
	if err != nil {
		return "", fmt.Errorf("unable to parse docker image: %s", err)
	}
	named, ok := ref.(reference.Named)
	if !ok {
		return "", fmt.Errorf("had an issue with the docker image")
	}
	return fmt.Sprintf("%s@%s", named.Name(), digest), nil
}

func (p *Porter) buildBundle(invocationImage string, digest string) error {
	fmt.Printf("\nGenerating Bundle File with Invocation Image %s =======> \n", invocationImage)
	bundle := Bundle{
		Name:    p.Config.Manifest.Name,
		Version: p.Config.Manifest.Version,
	}
	image := InvocationImage{
		Image:     invocationImage,
		ImageType: "docker",
	}
	image.Digest = digest
	bundle.InvocationImages = []InvocationImage{image}
	return p.WriteFile(bundle, 0644)

}
