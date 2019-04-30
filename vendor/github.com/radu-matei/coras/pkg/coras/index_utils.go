package coras

import (
	"fmt"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func getIndexFromImage(i *bundle.BaseImage) (ocispec.Descriptor, error) {
	im, err := getOCIImage(i.Image)
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("cannot get image: %v", err)
	}

	rm, err := im.RawManifest()
	if err != nil {
		return ocispec.Descriptor{}, fmt.Errorf("cannot get manifest: %v", err)
	}

	return ocispec.Descriptor{
		MediaType: ocispec.MediaTypeImageManifest,
		Size:      int64(len(rm)),
		Digest:    digest.FromBytes(rm),
		//Annotations: map[string]string{"io.cnab.manifest.type": "invocation"},
	}, nil
}

func getOCIImage(imageRef string) (v1.Image, error) {
	auth, err := resolve(imageRef)
	if err != nil {
		return nil, err
	}

	im, err := name.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("cannit parse image reference: %v", err)
	}

	return remote.Image(im, remote.WithAuth(auth))
}

func resolve(imageRef string) (authn.Authenticator, error) {
	// TODO - @radu-matei
	// support digest-referenced images
	repo, err := name.NewRepository(strings.Split(imageRef, ":")[0], name.WeakValidation)
	if err != nil {
		return nil, fmt.Errorf("cannot get repository name: %v", err)
	}

	return authn.DefaultKeychain.Resolve(repo.Registry)
}
