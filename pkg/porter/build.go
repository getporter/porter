package porter

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cxt "github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/pkg/errors"
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
		return errors.Wrap(err, "unable to build CNAB invocation image")
	}

	taggedImage, err := p.rewriteImageWithDigest(p.Config.Manifest.Image, digest)
	if err != nil {
		return fmt.Errorf("unable to regenerate tag: %s", err)
	}

	return p.buildBundle(taggedImage, digest)
}

func (p *Porter) generateDockerFile() error {
	lines, err := p.buildDockerFile()
	if err != nil {
		return errors.Wrap(err, "error generating the Dockerfile")
	}

	fmt.Fprintf(p.Out, "\nWriting Dockerfile =======>\n")
	contents := strings.Join(lines, "\n")
	fmt.Fprintln(p.Out, contents)
	err = p.Config.FileSystem.WriteFile("Dockerfile", []byte(contents), 0644)
	return errors.Wrap(err, "couldn't write the Dockerfile")
}

func (p *Porter) buildDockerFile() ([]string, error) {
	fmt.Fprintf(p.Out, "\nGenerating Dockerfile =======>\n")

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

	fmt.Fprintln(p.Out, lines)

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
	return `CMD ["/cnab/app/run"]`
}

func (p *Porter) buildMixinsSection() ([]string, error) {
	for _, m := range p.Manifest.Mixins {
		mixinDir, err := p.GetMixinDir(m)
		if err != nil {
			return nil, err
		}

		r := mixin.NewRunner(m, mixinDir, false)
		r.Command = "build"
		r.Data = "" // TODO: let the mixin know about which steps will be executed so that it can be more selective about copying into the invocation image

		// Copy the existing context and tweak to pipe the output differently
		mixinStdout := &bytes.Buffer{}
		var mixinContext cxt.Context
		mixinContext = *p.Context
		mixinContext.Out = mixinStdout   // mixin stdout -> dockerfile lines
		mixinContext.Err = p.Context.Out // mixin stderr -> logs
		r.Context = &mixinContext

		err = r.Validate()
		if err != nil {
			return nil, err
		}

		err = r.Run()
		if err != nil {
			return nil, err
		}
		return strings.Split(mixinStdout.String(), "\n"), nil
	}
	return nil, nil
}

func (p *Porter) copyMixins() error {
	fmt.Printf("Processing mixins ===> \n")
	for _, mixin := range append(p.Manifest.Mixins, "porter") {
		err := p.copyMixin(mixin)
		if err != nil {
			return err
		}
	}

	// Make the porter runtime available at the root of the app
	err := p.Context.CopyFile("cnab/app/mixins/porter/porter-runtime", "cnab/app/porter-runtime")
	return errors.Wrap(err, "could not copy porter-runtime mixin")
}

func (p *Porter) copyMixin(mixin string) error {
	fmt.Printf("Processing mixin %s ===> \n", mixin)
	mixinDir, _ := p.GetMixinDir(mixin)

	dirExists, err := p.FileSystem.DirExists(mixinDir)
	if err != nil {
		return errors.Wrapf(err, "could not check if directory exists %q", mixinDir)
	}
	if !dirExists {
		err := p.FileSystem.MkdirAll(mixinDir, 0755)
		if err != nil {
			return errors.Wrapf(err, "could not create mixin directory for %s", mixin)
		}
	}

	err = p.Context.CopyDirectory(mixinDir, "cnab/app/mixins", true)
	return errors.Wrapf(err, "could not copy mixin directory contents for %s", mixin)
}

func (p *Porter) buildInvocationImage(ctx context.Context) (string, error) {
	fmt.Printf("\nStarting Invocation Image Build =======> \n")
	path, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "could not get current working directory")
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
		return "", errors.Wrap(err, "cannot create Docker client")
	}
	cli.NegotiateAPIVersion(ctx)

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
		return "", errors.Wrap(err, "failed to stream docker build stdout")
	}

	authStr, err := p.getDockerAuth()
	if err != nil {
		return "", err
	}

	pushResponse, err := cli.ImagePush(ctx, p.Config.Manifest.Image, types.ImagePushOptions{
		All:          true,
		RegistryAuth: authStr,
	})
	if err != nil {
		return "", errors.Wrap(err, "docker push failed")
	}
	defer pushResponse.Close()
	err = jsonmessage.DisplayJSONMessagesStream(pushResponse, os.Stdout, termFd, isTerm, nil)
	if err != nil {
		if strings.HasPrefix(err.Error(), "denied") {
			return "", errors.Wrap(err, "docker push authentication failed")
		}
		return "", errors.Wrap(err, "failed to stream docker push stdout")
	}
	dist, err := cli.DistributionInspect(ctx, p.Config.Manifest.Image, authStr)
	if err != nil {
		return "", errors.Wrap(err, "unable to inspect docker image")
	}
	return string(dist.Descriptor.Digest), nil
}

func (p *Porter) getDockerAuth() (string, error) {
	authConfig := types.AuthConfig{
		Username: os.Getenv("DOCKER_USER"),
		Password: os.Getenv("DOCKER_PASSWORD"),
	}
	if authConfig.Username == "" || authConfig.Password == "" {
		return "", errors.New("DOCKER_USER and DOCKER_PASSWORD must be set")
	}

	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", errors.Wrap(err, "unable to build Docker auth")
	}
	return base64.URLEncoding.EncodeToString(encodedJSON), nil
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
