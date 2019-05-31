package porter

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/config"
	"github.com/docker/cli/cli/command"
	dockerconfig "github.com/docker/cli/cli/config"
	cliflags "github.com/docker/cli/cli/flags"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"

	"github.com/docker/cnab-to-oci/remotes"
	"github.com/pkg/errors"
)

// PublishOptions are options that may be specified when publishing a bundle.
// Porter handles defaulting any missing values.
type PublishOptions struct {
	File             string
	InsecureRegistry bool
}

// Validate performs validation on the publish options
func (p PublishOptions) Validate(porter *Porter) error {
	fs := porter.Context.FileSystem
	f := p.File
	if f == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "could not get current working directory")
		}
		f = filepath.Join(pwd, config.Name)

	}
	exists, err := fs.Exists(f)
	if err != nil {
		return errors.Wrap(err, "error finding porter.yaml")
	}
	if !exists {
		return errors.New("could not find porter.yaml. run `porter create` and `porter build` to create a new bundle before publishing")
	}
	return nil
}

// Publish is a composite function that publishes an invocation image, rewrites the porter manifest
// and then regenerates the bundle.json. Finally it [TODO] publishes the manifest to an OCI registry.
func (p *Porter) Publish(opts PublishOptions) error {
	var err error
	if opts.File != "" {
		err = p.Config.LoadManifestFrom(opts.File)
	} else {
		err = p.Config.LoadManifest()
	}
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

	b, err := p.Config.FileSystem.ReadFile("bundle.json")
	bun, err := bundle.ParseReader(bytes.NewBuffer(b))
	if err != nil {
		return errors.Wrap(err, "unable to load CNAB bundle")
	}

	if p.Config.Manifest.BundleTag == "" {
		return errors.New("porter.yaml must specify a `tag` value for this bundle")
	}

	ref, err := parseOCIReference(p.Config.Manifest.BundleTag) //tag from manifest
	if err != nil {
		return errors.Wrap(err, "invalid bundle tag reference. expected value is REGISTRY/bundle:tag")
	}
	insecureRegistries := []string{}
	if opts.InsecureRegistry {
		reg := reference.Domain(ref)
		fmt.Printf("Registry is: %s", reg)
		insecureRegistries = append(insecureRegistries, reg)
	}

	resolverConfig := createResolver(insecureRegistries)

	err = remotes.FixupBundle(context.Background(), &bun, ref, resolverConfig, remotes.WithEventCallback(displayEvent))
	if err != nil {
		return err
	}
	d, err := remotes.Push(context.Background(), &bun, ref, resolverConfig.Resolver, true)
	if err != nil {
		return err
	}
	fmt.Printf("Pushed successfully, with digest %q\n", d.Digest)
	return nil
}

func createResolver(insecureRegistries []string) remotes.ResolverConfig {
	return remotes.NewResolverConfigFromDockerConfigFile(dockerconfig.LoadDefaultConfigFile(os.Stderr), insecureRegistries...)
}

func displayEvent(ev remotes.FixupEvent) {
	switch ev.EventType {
	case remotes.FixupEventTypeCopyImageStart:
		fmt.Fprintf(os.Stderr, "Starting to copy image %s...\n", ev.SourceImage)
	case remotes.FixupEventTypeCopyImageEnd:
		if ev.Error != nil {
			fmt.Fprintf(os.Stderr, "Failed to copy image %s: %s\n", ev.SourceImage, ev.Error)
		} else {
			fmt.Fprintf(os.Stderr, "Completed image %s copy\n", ev.SourceImage)
		}
	}
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

	ref, err := parseOCIReference(p.Config.Manifest.Image)
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

func parseOCIReference(ociRef string) (reference.Named, error) {
	return reference.ParseNormalizedNamed(ociRef)
}
