package cnabtooci

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/cnabio/cnab-to-oci/remotes"
	containerdRemotes "github.com/containerd/containerd/remotes"
	"github.com/docker/cli/cli/command"
	dockerconfig "github.com/docker/cli/cli/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/moby/term"
	"github.com/opencontainers/go-digest"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"
)

// ErrNoContentDigest represents an error due to an image not having a
// corresponding content digest in a bundle definition
type ErrNoContentDigest error

// NewErrNoContentDigest returns an ErrNoContentDigest formatted with the
// provided image name
func NewErrNoContentDigest(image string) ErrNoContentDigest {
	return fmt.Errorf("unable to verify that the pulled image %s is the invocation image referenced by the bundle because the bundle does not specify a content digest. This could allow for the invocation image to be replaced or tampered with", image)
}

var _ RegistryProvider = &Registry{}

type Registry struct {
	*portercontext.Context
}

func NewRegistry(c *portercontext.Context) *Registry {
	return &Registry{
		Context: c,
	}
}

// PullBundle pulls a bundle from an OCI registry. Returns the bundle, and an optional image relocation mapping, if applicable.
func (r *Registry) PullBundle(ctx context.Context, ref cnab.OCIReference, insecureRegistry bool) (cnab.BundleReference, error) {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("reference", ref.String()),
		attribute.Bool("insecure", insecureRegistry),
	)
	defer span.EndSpan()

	var insecureRegistries []string
	if insecureRegistry {
		reg := ref.Registry()
		insecureRegistries = append(insecureRegistries, reg)
	}

	if span.ShouldLog(zapcore.DebugLevel) {
		msg := strings.Builder{}
		msg.WriteString("Pulling bundle ")
		msg.WriteString(ref.String())
		if insecureRegistry {
			msg.WriteString(" with --insecure-registry")
		}
		span.Debug(msg.String())
	}

	bun, reloMap, digest, err := remotes.Pull(ctx, ref.Named, r.createResolver(insecureRegistries))
	if err != nil {
		return cnab.BundleReference{}, fmt.Errorf("unable to pull bundle: %w", err)
	}

	invocationImage := bun.InvocationImages[0]
	if invocationImage.Digest == "" {
		return cnab.BundleReference{}, NewErrNoContentDigest(invocationImage.Image)
	}

	bundleRef := cnab.BundleReference{
		Reference:     ref,
		Digest:        digest,
		Definition:    cnab.NewBundle(*bun),
		RelocationMap: reloMap,
	}

	return bundleRef, nil
}

func (r *Registry) PushBundle(ctx context.Context, bundleRef cnab.BundleReference, insecureRegistry bool) (cnab.BundleReference, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var insecureRegistries []string
	if insecureRegistry {
		reg := bundleRef.Reference.Registry()
		insecureRegistries = append(insecureRegistries, reg)
	}

	resolver := r.createResolver(insecureRegistries)

	// Initialize the relocation map if necessary
	if bundleRef.RelocationMap == nil {
		bundleRef.RelocationMap = make(relocation.ImageRelocationMap)
	}
	rm, err := remotes.FixupBundle(context.Background(), &bundleRef.Definition.Bundle, bundleRef.Reference.Named, resolver, remotes.WithEventCallback(r.displayEvent), remotes.WithAutoBundleUpdate(), remotes.WithRelocationMap(bundleRef.RelocationMap))
	if err != nil {
		return cnab.BundleReference{}, log.Error(fmt.Errorf("error preparing the bundle with cnab-to-oci before pushing: %w", err))
	}
	bundleRef.RelocationMap = rm

	d, err := remotes.Push(ctx, &bundleRef.Definition.Bundle, rm, bundleRef.Reference.Named, resolver, true)
	if err != nil {
		return cnab.BundleReference{}, log.Error(fmt.Errorf("error pushing the bundle to %s: %w", bundleRef.Reference, err))
	}
	bundleRef.Digest = d.Digest

	fmt.Fprintf(r.Out, "Bundle tag %s pushed successfully, with digest %q\n", bundleRef.Reference, d.Digest)
	return bundleRef, nil
}

// PushInvocationImage pushes the invocation image from the Docker image cache to the specified location
// the expected format of the invocationImage is REGISTRY/NAME:TAG.
// Returns the image digest from the registry.
func (r *Registry) PushInvocationImage(ctx context.Context, invocationImage string) (digest.Digest, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	cli, err := docker.GetDockerClient()
	if err != nil {
		return "", err
	}

	ref, err := cnab.ParseOCIReference(invocationImage)
	if err != nil {
		return "", err
	}
	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := ref.ParseRepositoryInfo()
	if err != nil {
		return "", err
	}
	authConfig := command.ResolveAuthConfig(ctx, cli, repoInfo.Index)
	encodedAuth, err := command.EncodeAuthToBase64(authConfig)
	if err != nil {
		return "", err
	}
	options := types.ImagePushOptions{
		RegistryAuth: encodedAuth,
	}

	fmt.Fprintln(r.Out, "Pushing CNAB invocation image...")
	pushResponse, err := cli.Client().ImagePush(ctx, invocationImage, options)
	if err != nil {
		return "", fmt.Errorf("docker push failed: %w", err)
	}
	defer pushResponse.Close()

	termFd, _ := term.GetFdInfo(r.Out)
	// Setting this to false here because Moby os.Exit(1) all over the place and this fails on WSL (only)
	// when Term is true.
	isTerm := false
	err = jsonmessage.DisplayJSONMessagesStream(pushResponse, r.Out, termFd, isTerm, nil)
	if err != nil {
		if strings.HasPrefix(err.Error(), "denied") {
			return "", fmt.Errorf("docker push authentication failed: %w", err)
		}
		return "", fmt.Errorf("failed to stream docker push stdout: %w", err)
	}
	dist, err := cli.Client().DistributionInspect(ctx, invocationImage, encodedAuth)
	if err != nil {
		return "", fmt.Errorf("unable to inspect docker image: %w", err)
	}
	return dist.Descriptor.Digest, nil
}

func (r *Registry) createResolver(insecureRegistries []string) containerdRemotes.Resolver {
	return remotes.CreateResolver(dockerconfig.LoadDefaultConfigFile(r.Out), insecureRegistries...)
}

func (r *Registry) displayEvent(ev remotes.FixupEvent) {
	switch ev.EventType {
	case remotes.FixupEventTypeCopyImageStart:
		fmt.Fprintf(r.Out, "Starting to copy image %s...\n", ev.SourceImage)
	case remotes.FixupEventTypeCopyImageEnd:
		if ev.Error != nil {
			fmt.Fprintf(r.Out, "Failed to copy image %s: %s\n", ev.SourceImage, ev.Error)
		} else {
			fmt.Fprintf(r.Out, "Completed image %s copy\n", ev.SourceImage)
		}
	}
}

func (r *Registry) IsImageCached(ctx context.Context, invocationImage string) (bool, error) {
	cli, err := docker.GetDockerClient()
	if err != nil {
		return false, err
	}

	imageListOpts := types.ImageListOptions{All: true, Filters: filters.NewArgs(filters.KeyValuePair{Key: "reference", Value: invocationImage})}

	imageSummaries, err := cli.Client().ImageList(ctx, imageListOpts)
	if err != nil {
		return false, fmt.Errorf("could not list images: %w", err)
	}

	if len(imageSummaries) == 0 {
		return false, nil
	}

	return true, nil
}

func (r *Registry) ListTags(ctx context.Context, repository string) ([]string, error) {
	tags, err := crane.ListTags(repository)
	if err != nil {
		return nil, fmt.Errorf("error listing tags for %s: %w", repository, err)
	}

	return tags, nil
}
