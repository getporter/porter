package coras

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/cnab-go/bundle"
	auth "github.com/deislabs/oras/pkg/auth/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// CNABThinMediaType represents a *temporary* media type for thin CNAB bundles
//
// TODO - @radu-matei
// discuss media types for CNAB
const CNABThinMediaType = "application/vnd.cnab.bundle.thin.v1-wd+json"

// CNABThickMediaType represents a thick bundle
const CNABThickMediaType = "application/vnd.cnab.bundle.thick.v1-wd+tgz"

// CNABThinBundleFileName represents the name of a thin bundle as stored in the registry
const CNABThinBundleFileName = "bundle.json"

// CNABThickBundleFileName represents the name of a thick bundle as stored in the registry
const CNABThickBundleFileName = "bundle.tgz"

// Push pushes a bundle to an OCI registry
func Push(inputFile, targetRef string, exported bool) error {
	if exported {
		return pushThick(inputFile, targetRef)
	}

	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("cannot read input bundle: %v", err)
	}

	b, err := bundle.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("cannot unmarshal input bundle: %v", err)
	}

	return pushThin(b, targetRef)
}

// pushThin pushes a thin bundle and relocates all images to a new repository
func pushThin(b *bundle.Bundle, targetRef string) error {
	err := RelocateBundleImages(b, targetRef)
	if err != nil {
		return err
	}

	data, err := json.Marshal(b)
	if err != nil {
		return err
	}

	ms := content.NewMemoryStore()
	desc := ms.Add(CNABThinBundleFileName, CNABThinMediaType, data)
	pushContents := []ocispec.Descriptor{desc}
	_, err = oras.Push(context.Background(), newResolver(), targetRef, ms, pushContents)

	return err
}

// pushThick pushes a thick bundle to an OCI registry.
//
// The resulting image will have a single layer, the bundle archive .tgz file
func pushThick(archiveFile string, targetRef string) error {
	data, err := ioutil.ReadFile(archiveFile)
	if err != nil {
		return fmt.Errorf("cannot read exported bundle: %v", err)
	}

	ms := content.NewMemoryStore()
	desc := ms.Add(CNABThickBundleFileName, CNABThickMediaType, data)
	pushContents := []ocispec.Descriptor{desc}
	_, err = oras.Push(context.Background(), newResolver(), targetRef, ms, pushContents)

	return err
}

func newResolver() remotes.Resolver {
	cli, err := auth.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error loading auth file: %v\n", err)
	}

	resolver, err := cli.Resolver(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Error loading resolver: %v\n", err)
		resolver = docker.NewResolver(docker.ResolverOptions{})
	}

	return resolver
}
