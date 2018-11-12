package porter

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
)

func (p *Porter) Build() error {
	p.Config.LoadManifest("porter.yaml")
	ctx := context.Background()
	if err := p.generateDockerFile(); err != nil {
		return fmt.Errorf("unable to generate Dockerfile: %s", err)
	}
	digest, err := p.buildInvocationImage(ctx)
	if err != nil {
		return fmt.Errorf("unable to build CNAB invocation image: %s", err)
	}
	return p.buildBundle(p.Config.Manifest.Image, digest)
}

func (p *Porter) generateDockerFile() error {
	f, err := p.Config.FileSystem.OpenFile("Dockerfile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open Dockerfile: %s", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	err = p.addDockerBaseImage(w)
	err = p.addCNAB(w)
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

func (p *Porter) addCNAB(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "ADD cnab/ cnab/\n"); err != nil {
		return fmt.Errorf("couldn't write docker base image: %s", err)
	}
	return nil
}

func (p *Porter) addMixins(w io.Writer) error {
	copyTemplate := "COPY %s cnab/app/%s\n"
	// Always copy the porter and exec "mixins"
	homedir, _ := p.GetHomeDir()
	mixinDir := fmt.Sprintf("%s/mixins", homedir)
	porterPath := fmt.Sprintf("%s/%s", mixinDir, "porter")
	porterMixin, _ := p.Config.FileSystem.Open(porterPath)
	defer porterMixin.Close()
	f, _ := p.Config.FileSystem.OpenFile("porter", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer f.Close()
	io.Copy(porterMixin, f)
	if _, err := fmt.Fprintf(w, fmt.Sprintf(copyTemplate, "porter", "porter")); err != nil {
		return fmt.Errorf("couldn't write docker base image: %s", err)
	}

	execPath := fmt.Sprintf("%s/%s", mixinDir, "exec")
	execMixin, _ := p.Config.FileSystem.Open(execPath)
	defer execMixin.Close()
	f, _ = p.Config.FileSystem.OpenFile("exec", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer f.Close()
	io.Copy(execMixin, f)
	if _, err := fmt.Fprintf(w, fmt.Sprintf(copyTemplate, "exec", "mixins/exec")); err != nil {
		return fmt.Errorf("couldn't write docker base image: %s", err)
	}

	for _, mixin := range p.Manifest.Mixins {
		mixinExec, _ := p.Config.FileSystem.Open(fmt.Sprintf("%s/%s", mixinDir, mixin))
		defer execMixin.Close()
		f, _ = p.Config.FileSystem.OpenFile(mixin, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		defer f.Close()
		io.Copy(mixinExec, f)
		if _, err := fmt.Fprintf(w, fmt.Sprintf(copyTemplate, mixin, "mixins/"+mixin)); err != nil {
			return fmt.Errorf("couldn't write docker base image: %s", err)
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
	fmt.Printf("Starting Invocation Image Build")
	path, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error %s", err)
	}
	fmt.Println(path)
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
		panic(err)
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

func (p *Porter) buildBundle(invocationImage string, digest string) error {
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
