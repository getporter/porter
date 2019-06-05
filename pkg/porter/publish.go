package porter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deislabs/porter/pkg/config"
	"github.com/docker/cli/cli/command"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	"github.com/pkg/errors"
)

// PublishOptions are options that may be specified when publishing a bundle.
// Porter handles defaulting any missing values.
type PublishOptions struct {
	bundleFileOptions
	Tag string
}

func (p PublishOptions) Validate(porter *Porter) error {
	if p.File == "" {
		fs := porter.Context.FileSystem
		pwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "could not get current working directory")
		}

		exists, err := fs.Exists(filepath.Join(pwd, config.Name))
		if err != nil {
			return errors.Wrap(err, "error finding porter.yaml")
		}

		if !exists {
			return errors.New("could not find porter.yaml. run `porter create` and `porter build` to create a new bundle before publishing")
		}
	}

	// Bundle tags must be valid OCI reference formats.
	if p.Tag != "" {
		_, err := reference.ParseNormalizedNamed(p.Tag)
		if err != nil {
			return errors.Wrap(err, "invalid --tag value. expected format is REGISTRY/IMAGE:TAG")
		}
	}
	return nil
}

// Publish is a composite function that publishes an invocation image, rewrites the porter manifest
// and then regenerates the bundle.json. Finally it [TODO] publishes the manifest to an OCI registry.
func (p *Porter) Publish(opts PublishOptions) error {
	var err error
	if opts.File != "" { // TODO: Extract validation from sharedOptions so that we aren't diverging logic from the other commands like we are here. Normally file is always populated by Validate.
		err = p.Config.LoadManifestFrom(opts.File)
	} else {
		err = p.Config.LoadManifest()
	}
	if err != nil {
		return err
	}

	err = p.EnsureBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	ctx := context.Background()
	cli, err := p.getDockerClient(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintln(p.Out, "Pushing CNAB invocation image...")
	digest, err := p.publishInvocationImage(ctx, cli)
	if err != nil {
		return errors.Wrap(err, "unable to push CNAB invocation image")
	}

	taggedImage, err := p.rewriteImageWithDigest(p.Config.Manifest.Image, digest)
	if err != nil {
		return errors.Wrap(err, "unable to update invocation image reference: %s")
	}

	fmt.Fprintln(p.Out, "\nGenerating CNAB bundle.json...")
	err = p.buildBundle(taggedImage, digest)
	if err != nil {
		return errors.Wrap(err, "unable to generate CNAB bundle.json")
	}

	if opts.Tag != "" {
		//p.Config.Manifest.Tag = opts.Tag
	}
	// TODO: uncomment this when cnab-to-oci is integrated
	//fmt.Fprintf(p.Out, "Tagging bundle image as %s...\n", "")

	// TODO: Use CNAB-to-OCI to push the bundle (see https://github.com/deislabs/porter/issues/254)
	return nil
}

func (p *Porter) getDockerClient(ctx context.Context) (*command.DockerCli, error) {
	cli, err := command.NewDockerCli()
	if err != nil {
		return nil, errors.Wrap(err, "could not create new docker client")
	}
	if err := cli.Initialize(cliflags.NewClientOptions()); err != nil {
		return nil, err
	}
	return cli, nil
}

func (p *Porter) publishInvocationImage(ctx context.Context, cli *command.DockerCli) (string, error) {

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

	termFd, _ := term.GetFdInfo(os.Stdout)
	// Setting this to false here because Moby os.Exit(1) all over the place and this fails on WSL (only)
	// when Term is true.
	isTerm := false
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
