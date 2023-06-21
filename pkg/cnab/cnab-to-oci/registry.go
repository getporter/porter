package cnabtooci

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
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
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
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
func (r *Registry) PullBundle(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (cnab.BundleReference, error) {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("reference", ref.String()),
		attribute.Bool("insecure", opts.InsecureRegistry),
	)
	defer span.EndSpan()

	var insecureRegistries []string
	if opts.InsecureRegistry {
		reg := ref.Registry()
		insecureRegistries = append(insecureRegistries, reg)
	}
	resolver := r.createResolver(insecureRegistries)

	if span.ShouldLog(zapcore.DebugLevel) {
		msg := strings.Builder{}
		msg.WriteString("Pulling bundle ")
		msg.WriteString(ref.String())
		if opts.InsecureRegistry {
			msg.WriteString(" with --insecure-registry")
		}
		span.Debug(msg.String())
	}

	bun, reloMap, digest, err := remotes.Pull(ctx, ref.Named, resolver)
	if err != nil {
		return cnab.BundleReference{}, span.Errorf("unable to pull bundle: %w", err)
	}

	invocationImage := bun.InvocationImages[0]
	if invocationImage.Digest == "" {
		return cnab.BundleReference{}, span.Error(NewErrNoContentDigest(invocationImage.Image))
	}

	bundleRef := cnab.BundleReference{
		Reference:     ref,
		Digest:        digest,
		Definition:    cnab.NewBundle(*bun),
		RelocationMap: reloMap,
	}

	return bundleRef, nil
}

func (r *Registry) PushBundle(ctx context.Context, bundleRef cnab.BundleReference, opts RegistryOptions) (cnab.BundleReference, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var insecureRegistries []string
	if opts.InsecureRegistry {
		// Get all source registries
		registries, err := bundleRef.Definition.GetReferencedRegistries()
		if err != nil {
			return cnab.BundleReference{}, err
		}

		// Include our destination registry
		destReg := bundleRef.Reference.Registry()
		found := false
		for _, reg := range registries {
			if destReg == reg {
				found = true
			}
		}
		if !found {
			registries = append(registries, destReg)
		}

		// All registries used should be marked as allowing insecure connections
		insecureRegistries = registries
		log.SetAttributes(attribute.String("insecure-registries", strings.Join(registries, ",")))
	}
	resolver := r.createResolver(insecureRegistries)

	if log.ShouldLog(zapcore.DebugLevel) {
		msg := strings.Builder{}
		msg.WriteString("Pushing bundle ")
		msg.WriteString(bundleRef.String())
		if opts.InsecureRegistry {
			msg.WriteString(" with --insecure-registry")
		}
		log.Debug(msg.String())
	}

	// Initialize the relocation map if necessary
	if bundleRef.RelocationMap == nil {
		bundleRef.RelocationMap = make(relocation.ImageRelocationMap)
	}
	rm, err := remotes.FixupBundle(ctx, &bundleRef.Definition.Bundle, bundleRef.Reference.Named, resolver,
		remotes.WithEventCallback(r.displayEvent),
		remotes.WithAutoBundleUpdate(),
		remotes.WithRelocationMap(bundleRef.RelocationMap))
	if err != nil {
		return cnab.BundleReference{}, log.Error(fmt.Errorf("error preparing the bundle with cnab-to-oci before pushing: %w", err))
	}
	bundleRef.RelocationMap = rm

	d, err := remotes.Push(ctx, &bundleRef.Definition.Bundle, rm, bundleRef.Reference.Named, resolver, true)
	if err != nil {
		return cnab.BundleReference{}, log.Error(fmt.Errorf("error pushing the bundle to %s: %w", bundleRef.Reference, err))
	}
	bundleRef.Digest = d.Digest

	log.Infof("Bundle %s pushed successfully, with digest %q\n", bundleRef.Reference, d.Digest)
	return bundleRef, nil
}

// PushImage pushes the image from the Docker image cache to the specified location
// the expected format of the image is REGISTRY/NAME:TAG.
// Returns the image digest from the registry.
func (r *Registry) PushImage(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (digest.Digest, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	cli, err := docker.GetDockerClient()
	if err != nil {
		return "", log.Errorf("error creating a docker client: %w", err)
	}

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := ref.ParseRepositoryInfo()
	if err != nil {
		return "", log.Errorf("error parsing the repository potion of the image reference %s: %w", ref, err)
	}
	authConfig := command.ResolveAuthConfig(ctx, cli, repoInfo.Index)
	encodedAuth, err := command.EncodeAuthToBase64(authConfig)
	if err != nil {
		return "", log.Errorf("error encoding authentication information for the docker client: %w", err)
	}
	options := types.ImagePushOptions{
		RegistryAuth: encodedAuth,
	}

	log.Info("Pushing bundle image...")
	pushResponse, err := cli.Client().ImagePush(ctx, ref.String(), options)
	if err != nil {
		return "", log.Errorf("docker push failed: %w", err)
	}
	defer pushResponse.Close()

	termFd, _ := term.GetFdInfo(r.Out)
	// Setting this to false here because Moby os.Exit(1) all over the place and this fails on WSL (only)
	// when Term is true.
	isTerm := false
	err = jsonmessage.DisplayJSONMessagesStream(pushResponse, r.Out, termFd, isTerm, nil)
	if err != nil {
		if strings.HasPrefix(err.Error(), "denied") {
			return "", log.Errorf("docker push authentication failed: %w", err)
		}
		return "", log.Errorf("failed to stream docker push stdout: %w", err)
	}
	dist, err := cli.Client().DistributionInspect(ctx, ref.String(), encodedAuth)
	if err != nil {
		return "", log.Errorf("unable to inspect docker image: %w", err)
	}
	return dist.Descriptor.Digest, nil
}

// PullImage pulls an image from an OCI registry.
func (r *Registry) PullImage(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	cli, err := docker.GetDockerClient()
	if err != nil {
		return log.Error(err)
	}

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := ref.ParseRepositoryInfo()
	if err != nil {
		return log.Error(err)
	}
	authConfig := command.ResolveAuthConfig(ctx, cli, repoInfo.Index)
	encodedAuth, err := command.EncodeAuthToBase64(authConfig)
	if err != nil {
		return log.Error(fmt.Errorf("failed to serialize docker auth config: %w", err))
	}
	options := types.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}

	imgRef := ref.String()
	rd, err := cli.Client().ImagePull(ctx, imgRef, options)
	if err != nil {
		return log.Error(fmt.Errorf("docker pull for image %s failed: %w", imgRef, err))
	}
	defer rd.Close()

	// save the image to docker cache
	_, err = io.ReadAll(rd)
	if err != nil {
		return fmt.Errorf("failed to save image %s into local cache: %w", imgRef, err)
	}

	return nil
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

// GetCachedImage returns information about an image from local docker cache.
func (r *Registry) GetCachedImage(ctx context.Context, ref cnab.OCIReference) (ImageMetadata, error) {
	image := ref.String()
	ctx, log := tracing.StartSpan(ctx, attribute.String("reference", image))
	defer log.EndSpan()

	cli, err := docker.GetDockerClient()
	if err != nil {
		return ImageMetadata{}, log.Error(err)
	}

	result, _, err := cli.Client().ImageInspectWithRaw(ctx, image)
	if err != nil {
		err = fmt.Errorf("failed to find image in docker cache: %w", ErrNotFound{Reference: ref})
		// log as debug because this isn't a terminal error
		log.Debugf(err.Error())
		return ImageMetadata{}, err
	}

	summary, err := NewImageSummaryFromInspect(ref, result)
	if err != nil {
		return ImageMetadata{}, log.Error(fmt.Errorf("failed to extract image %s in docker cache: %w", image, err))
	}

	return summary, nil
}

func (r *Registry) ListTags(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) ([]string, error) {
	// Get the fully-qualified repository name, including docker.io (required by crane)
	repository := ref.Named.Name()

	//lint:ignore SA4006 ignore unused context for now
	ctx, span := tracing.StartSpan(ctx, attribute.String("repository", repository))
	defer span.EndSpan()

	tags, err := crane.ListTags(repository, opts.toCraneOptions()...)
	if err != nil {
		if notFoundErr := asNotFoundError(err, ref); notFoundErr != nil {
			return nil, span.Error(notFoundErr)
		}
		return nil, span.Errorf("error listing tags for %s: %w", ref.String(), err)
	}

	return tags, nil
}

// GetBundleMetadata returns information about a bundle in a registry
// Use ErrNotFound to detect if the error is because the bundle is not in the registry.
func (r *Registry) GetBundleMetadata(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (BundleMetadata, error) {
	//lint:ignore SA4006 ignore unused context for now
	ctx, span := tracing.StartSpan(ctx, attribute.String("reference", ref.String()))
	defer span.EndSpan()

	bundleDigest, err := crane.Digest(ref.String(), opts.toCraneOptions()...)
	if err != nil {
		if notFoundErr := asNotFoundError(err, ref); notFoundErr != nil {
			return BundleMetadata{}, span.Error(notFoundErr)
		}
		return BundleMetadata{}, span.Errorf("error retrieving bundle metadata for %s: %w", ref.String(), err)
	}

	return BundleMetadata{
		BundleReference: cnab.BundleReference{
			Reference: ref,
			Digest:    digest.Digest(bundleDigest),
		},
	}, nil
}

// GetImageMetadata returns information about an image in a registry
// Use ErrNotFound to detect if the error is because the image is not in the registry.
func (r *Registry) GetImageMetadata(ctx context.Context, ref cnab.OCIReference, opts RegistryOptions) (ImageMetadata, error) {
	ctx, span := tracing.StartSpan(ctx, attribute.String("reference", ref.String()))
	defer span.EndSpan()

	// Check if we already have the image in the Docker cache
	cachedResult, err := r.GetCachedImage(ctx, ref)
	if err != nil {
		if !errors.Is(err, ErrNotFound{}) {
			return ImageMetadata{}, err
		}
	}

	// Check if we have the repository digest cached for the referenced image
	if cachedDigest, err := cachedResult.GetRepositoryDigest(); err == nil {
		span.SetAttributes(attribute.String("cached-digest", cachedDigest.String()))
		return cachedResult, nil
	}

	// Do a HEAD against the registry to retrieve image metadata without pulling the entire image contents
	desc, err := crane.Head(ref.String(), opts.toCraneOptions()...)
	if err != nil {
		if notFoundErr := asNotFoundError(err, ref); notFoundErr != nil {
			return ImageMetadata{}, span.Error(notFoundErr)
		}
		return ImageMetadata{}, span.Errorf("error fetching image metadata for %s: %w", ref, err)
	}

	repoDigest := digest.NewDigestFromHex(desc.Digest.Algorithm, desc.Digest.Hex)
	span.SetAttributes(attribute.String("fetched-digest", repoDigest.String()))

	return NewImageSummaryFromDigest(ref, repoDigest)
}

// asNotFoundError checks if the error is an HTTP 404 not found error, and if so returns a corresponding ErrNotFound instance.
func asNotFoundError(err error, ref cnab.OCIReference) error {
	var httpError *transport.Error
	if errors.As(err, &httpError) {
		if httpError.StatusCode == http.StatusNotFound {
			return ErrNotFound{Reference: ref}
		}
	}

	return nil
}

// ImageMetadata contains information about an OCI image.
type ImageMetadata struct {
	Reference   cnab.OCIReference
	RepoDigests []string
}

func NewImageSummaryFromInspect(ref cnab.OCIReference, sum types.ImageInspect) (ImageMetadata, error) {
	img := ImageMetadata{
		Reference:   ref,
		RepoDigests: sum.RepoDigests,
	}
	if img.IsZero() {
		return ImageMetadata{}, fmt.Errorf("invalid image summary for image reference %s", ref)
	}

	return img, nil
}

func NewImageSummaryFromDigest(ref cnab.OCIReference, repoDigest digest.Digest) (ImageMetadata, error) {
	digestedRef, err := ref.WithDigest(repoDigest)
	if err != nil {
		return ImageMetadata{}, fmt.Errorf("error building an OCI reference from image %s and digest %s", ref.Repository(), ref.Digest())
	}

	return ImageMetadata{
		Reference:   ref,
		RepoDigests: []string{digestedRef.String()},
	}, nil
}

func (i ImageMetadata) String() string {
	return i.Reference.String()
}

func (i ImageMetadata) IsZero() bool {
	return i.String() == ""
}

// GetRepositoryDigest finds the repository digest associated with the original
// image reference used to create this ImageMetadata.
func (i ImageMetadata) GetRepositoryDigest() (digest.Digest, error) {
	if len(i.RepoDigests) == 0 {
		return "", fmt.Errorf("failed to get digest for image: %s", i)
	}
	var imgDigest digest.Digest
	for _, rd := range i.RepoDigests {
		imgRef, err := cnab.ParseOCIReference(rd)
		if err != nil {
			return "", err
		}
		if imgRef.Repository() != i.Reference.Repository() {
			continue
		}

		if !imgRef.HasDigest() {
			return "", fmt.Errorf("image summary does not contain digest for image: %s", imgRef.String())
		}

		imgDigest = imgRef.Digest()
		break
	}

	if imgDigest == "" {
		return "", fmt.Errorf("cannot find image digest for desired repo %s", i)
	}

	if err := imgDigest.Validate(); err != nil {
		return "", err
	}

	return imgDigest, nil
}

// GetInsecureRegistryTransport returns a copy of the default http transport
// with InsecureSkipVerify set so that we can use it with insecure registries.
func GetInsecureRegistryTransport() *http.Transport {
	skipTLS := http.DefaultTransport.(*http.Transport)
	skipTLS = skipTLS.Clone()
	skipTLS.TLSClientConfig.InsecureSkipVerify = true
	return skipTLS
}
