package porter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
)

const (
	mixinDirTemplate  = "cnab/app/mixins/%s"
	mixinExecTemplate = "cnab/app/mixins/%s/%s"
)

func (p *Porter) Build() error {
	err := p.Config.LoadManifest("porter.yaml")
	if err != nil {
		return nil
	}

	if err := p.copyMixins(); err != nil {
		return fmt.Errorf("unable to copy mixins: %s", err)
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
	err := p.copyMixins()
	if err != nil {
		return err
	}

	lines, err := p.buildDockerFile()
	if err != nil {
		return errors.Wrap(err, "error generating the Dockerfile")
	}

	fmt.Printf("\nWriting Dockerfile =======>\n")
	contents := strings.Join(lines, "\n")
	err = p.Config.FileSystem.WriteFile("Dockerfile", []byte(contents), 0644)
	return errors.Wrap(err, "couldn't write the Dockerfile")
}

func (p *Porter) buildDockerFile() ([]string, error) {
	fmt.Printf("\nGenerating Dockerfile =======>\n")

	lines := make([]string, 0, 10)

	lines = append(lines, p.buildFromSection())
	lines = append(lines, p.buildCNABSection()...)
	lines = append(lines, p.buildPorterSection()...)
	lines = append(lines, p.buildCMDSection())

	mixinLines, err := p.buildMixinsSection()
	if err != nil {
		return nil, errors.Wrap(err, "error generating Dockefile content for mixins")
	}

	lines = append(lines, mixinLines...)

	return lines, nil
}

func (p *Porter) buildFromSection() string {
	return `FROM ubuntu:latest`
}

func (p *Porter) buildPorterSection() []string {
	return []string{
		`COPY porter.yaml /cnab/app/porter.yaml`,
	}
}

func (p *Porter) buildCNABSection() []string {
	return []string{
		`COPY cnab/ /cnab/`,
	}
}

func (p *Porter) buildCMDSection() string {
	return `CMD [/cnab/app/run]`
}

func (p *Porter) buildMixinsSection() ([]string, error) {
	return nil, nil
}

func (p *Porter) copyMixins() error {

	// Always copy in porter
	porterPath, _ := p.GetMixinPath("porter")
	porterMixin, err := p.Config.FileSystem.ReadFile(porterPath)
	if err != nil {
		return errors.Wrapf(err, "couldn't read the porter binary for container build")
	}

	err = p.Config.FileSystem.WriteFile("cnab/app/porter", porterMixin, 0755)
	if err != nil {
		return errors.Wrap(err, "couldn't write the porter binary for container build")
	}

	err = p.copyMixin("porter")
	if err != nil {
		return err
	}

	cnabMixins := "cnab/app/mixins"
	mixinsDirExists, err := p.Config.FileSystem.DirExists(cnabMixins)
	if err != nil {
		return errors.Wrap(err, "couldn't verify mixins directory")
	}
	if !mixinsDirExists {
		p.Config.FileSystem.Mkdir(cnabMixins, 0755)
	}
	fmt.Printf("Processing mixins ===> \n")
	for _, mixin := range p.Manifest.Mixins {
		err := p.copyMixin(mixin)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Porter) copyMixin(mixin string) error {
	mixinPath, _ := p.GetMixinPath(mixin)

	fmt.Printf("Processing mixin %s ===> \n", mixin)
	fmt.Printf("Reading: %s\n", mixinPath)
	fmt.Printf("Writing: %s\n", fmt.Sprintf(mixinExecTemplate, mixin, mixin))

	mixinsDirExists, err := p.Config.FileSystem.DirExists(fmt.Sprintf(mixinDirTemplate, mixin))
	if err != nil {
		return fmt.Errorf("couldn't verify mixins directory: %s", err)
	}

	if !mixinsDirExists {
		p.Config.FileSystem.Mkdir(fmt.Sprintf(mixinDirTemplate, mixin), 0755)
	}

	mixinExec, err := p.Config.FileSystem.Open(mixinPath)
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
	bundle.Parameters = p.generateBundleParameters()
	return p.WriteFile(bundle, 0644)

}

func (p *Porter) generateBundleParameters() map[string]ParameterDefinition {
	params := map[string]ParameterDefinition{}
	for _, param := range p.Manifest.Parameters {
		fmt.Printf("Generating parameter definition %s ====>\n", param.Name)
		p := ParameterDefinition{
			DataType:      param.DataType,
			DefaultValue:  param.DefaultValue,
			AllowedValues: param.AllowedValues,
			Required:      param.Required,
			MinValue:      param.MinValue,
			MaxValue:      param.MaxValue,
			MinLength:     param.MinLength,
			MaxLength:     param.MaxLength,
		}
		if param.Metadata.Description != "" {
			p.Metadata = ParameterMetadata{Description: param.Metadata.Description}
		}
		if param.Destination != nil {
			p.Destination = &Location{
				EnvironmentVariable: param.Destination.EnvironmentVariable,
				Path:                param.Destination.Path,
			}
		} else {
			p.Destination = &Location{
				EnvironmentVariable: strings.ToUpper(param.Name),
			}
		}
		params[param.Name] = p
	}
	return params
}
