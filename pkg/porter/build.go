package porter

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	cxt "github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	"github.com/pkg/errors"
)

func (p *Porter) Build() error {
	err := p.Config.LoadManifest()
	if err != nil {
		return err
	}

	if err := p.prepareDockerFilesystem(); err != nil {
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

	lines = append(lines, p.buildFromSection()...)
	lines = append(lines, p.buildCNABSection()...)
	lines = append(lines, p.buildPorterSection()...)
	lines = append(lines, p.buildCMDSection())
	lines = append(lines, p.buildCopySSL())

	mixinLines, err := p.buildMixinsSection()
	if err != nil {
		return nil, errors.Wrap(err, "error generating Dockefile content for mixins")
	}
	lines = append(lines, mixinLines...)

	fmt.Fprintln(p.Out, lines)

	return lines, nil
}

func (p *Porter) buildFromSection() []string {
	return []string{
		`FROM quay.io/deis/lightweight-docker-go:v0.2.0`,
		`FROM debian:stretch`,
	}
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

func (p *Porter) buildCopySSL() string {
	return `COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt`
}

func (p *Porter) buildMixinsSection() ([]string, error) {
	lines := make([]string, 0)
	for _, m := range p.Manifest.Mixins {
		mixinDir, err := p.GetMixinDir(m)
		if err != nil {
			return nil, err
		}

		r := mixin.NewRunner(m, mixinDir, false)
		r.Command = "build"
		r.Step = "" // TODO: let the mixin know about which steps will be executed so that it can be more selective about copying into the invocation image

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

		l := strings.Split(mixinStdout.String(), "\n")
		lines = append(lines, l...)
	}
	return lines, nil
}

func (p *Porter) prepareDockerFilesystem() error {
	fmt.Printf("Copying dependencies ===> \n")
	for _, dep := range p.Manifest.Dependencies {
		err := p.copyDependency(dep.Name)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Copying mixins ===> \n")
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

func (p *Porter) copyDependency(bundle string) error {
	fmt.Printf("Copying bundle dependency %s ===> \n", bundle)
	bundleDir, err := p.GetBundleDir(bundle)
	if err != nil {
		return err
	}

	err = p.Context.CopyDirectory(bundleDir, "cnab/app/bundles", true)
	return errors.Wrapf(err, "could not copy bundle directory contents for %s", bundle)
}

func (p *Porter) copyMixin(mixin string) error {
	fmt.Printf("Copying mixin %s ===> \n", mixin)
	mixinDir, err := p.GetMixinDir(mixin)
	if err != nil {
		return err
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

	cli := command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false, nil)
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return "", err
	}

	response, err := cli.Client().ImageBuild(context.Background(), tar, buildOptions)
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

	ref, err := reference.ParseNormalizedNamed(p.Config.Manifest.Image)
	if err != nil {
		return "", err
	}

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return "", err
	}
	authConfig := command.ResolveAuthConfig(ctx, cli, repoInfo.Index)
	encodedAuth, err := command.EncodeAuthToBase64(authConfig)
	if err != nil {
		return "", err
	}
	options := types.ImagePushOptions{
		All:          true,
		RegistryAuth: encodedAuth,
	}

	pushResponse, err := cli.Client().ImagePush(ctx, p.Config.Manifest.Image, options)
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
	dist, err := cli.Client().DistributionInspect(ctx, p.Config.Manifest.Image, encodedAuth)
	if err != nil {
		return "", errors.Wrap(err, "unable to inspect docker image")
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
		Name:        p.Config.Manifest.Name,
		Description: p.Config.Manifest.Description,
		Version:     p.Config.Manifest.Version,
	}
	image := InvocationImage{
		Image:     invocationImage,
		ImageType: "docker",
	}
	image.Digest = digest
	bundle.InvocationImages = []InvocationImage{image}
	bundle.Parameters = p.generateBundleParameters()
	bundle.Credentials = p.generateBundleCredentials()
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

func (p *Porter) generateBundleCredentials() map[string]Location {
	params := map[string]Location{}
	for _, cred := range p.Manifest.Credentials {
		fmt.Printf("Generating credential %s ====>\n", cred.Name)
		l := Location{
			Path:                cred.Path,
			EnvironmentVariable: cred.EnvironmentVariable,
		}
		params[cred.Name] = l
	}
	return params
}
