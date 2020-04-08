package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"path/filepath"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/docker/cnab-to-oci/relocation"
	"github.com/pkg/errors"
)

// CachedBundle represents a bundle pulled from a registry that has been cached to
// the filesystem.
type CachedBundle struct {
	// cacheDir is the cache directory for the bundle (not the general cache directory)
	cacheDir string

	// Tag of the cached bundle.
	Tag string

	// Bundle is the cached bundle definition.
	Bundle bundle.Bundle

	// BundlePath is the location of the bundle.json in the cache.
	BundlePath string

	// Manifest is the optional porter.yaml manifest. This is only populated
	// when the bundle was a porter built bundle that had a manifest embedded in
	// the custom metadata.
	Manifest *manifest.Manifest

	// ManifestPath is the optional location of the porter.yaml in the cache.
	ManifestPath string

	// RelocationMap is the optional relocation map enclosed in the bundle.
	RelocationMap *relocation.ImageRelocationMap

	// RelocationFilePath is the optional location of the relocation file in the cache.
	RelocationFilePath string
}

// GetBundleID is the unique ID of the cached bundle.
func (cb *CachedBundle) GetBundleID() string {
	// hash the tag, tags have characters that won't work as part of a path
	// so hashing here to get a path friendly name
	bid := md5.Sum([]byte(cb.Tag))
	return hex.EncodeToString(bid[:])
}

// SetCacheDir sets the bundle specific cache directory based on the given Porter cache directory.
func (cb *CachedBundle) SetCacheDir(porterCacheDir string) {
	cb.cacheDir = filepath.Join(porterCacheDir, cb.GetBundleID())
}

// BuildBundlePath generates the potential location of the bundle.json, if it existed.
func (cb *CachedBundle) BuildBundlePath() string {
	return filepath.Join(cb.cacheDir, "cnab", "bundle.json")
}

// BuildRelocationFilePath generates the potential location of the relocation file, if it existed.
func (cb *CachedBundle) BuildRelocationFilePath() string {
	return filepath.Join(cb.cacheDir, "cnab", "relocation-mapping.json")
}

// BuildManifestPath generates the potential location of the manifest, if it existed.
func (cb *CachedBundle) BuildManifestPath() string {
	return filepath.Join(cb.cacheDir, config.Name)
}

// Load starts from the bundle tag, and hydrates the cached bundle from the cache.
func (cb *CachedBundle) Load(cxt *context.Context) (bool, error) {
	// Check that the bundle exists
	cb.BundlePath = cb.BuildBundlePath()
	bundleExists, err := cxt.FileSystem.Exists(cb.BundlePath)
	if err != nil {
		return false, errors.Wrapf(err, "unable to read bundle %s at %s", cb.Tag, cb.BundlePath)
	}
	if !bundleExists {
		return false, nil
	}

	// Check for the optional relocation mapping next to it
	reloPath := cb.BuildRelocationFilePath()
	reloExists, err := cxt.FileSystem.Exists(reloPath)
	if err != nil {
		return true, errors.Wrapf(err, "unable to read relocation mapping %s at %s", cb.Tag, reloPath)
	}
	if reloExists {
		cb.RelocationFilePath = reloPath
	}

	// Check for the optional manifest
	manifestPath := cb.BuildManifestPath()
	manifestExists, err := cxt.FileSystem.Exists(manifestPath)
	if err != nil {
		return true, errors.Wrapf(err, "unable to read manifest %s at %s", cb.Tag, manifestPath)
	}
	if manifestExists {
		cb.ManifestPath = manifestPath
	}

	// Read the files
	data, err := cxt.FileSystem.ReadFile(cb.BundlePath)
	if err != nil {
		return true, errors.Wrapf(err, "unable to read cached bundle file at %s", cb.BundlePath)
	}

	bun, err := bundle.Unmarshal(data)
	if err != nil {
		return true, errors.Wrapf(err, "unable to parse cached bundle file at %s", cb.BundlePath)
	}
	cb.Bundle = *bun

	if cb.RelocationFilePath != "" {
		data, err = cxt.FileSystem.ReadFile(cb.RelocationFilePath)
		if err != nil {
			return true, errors.Wrapf(err, "unable to read cached relocation file at %s", cb.RelocationFilePath)
		}

		reloMap := &relocation.ImageRelocationMap{}
		err = json.Unmarshal(data, reloMap)
		if err != nil {
			return true, errors.Wrapf(err, "unable to parse cached relocation file at %s", cb.RelocationFilePath)
		}
		cb.RelocationMap = reloMap
	}

	if cb.ManifestPath != "" {
		m, err := manifest.LoadManifestFrom(cxt, cb.ManifestPath)
		if err != nil {
			return true, errors.Wrapf(err, "unable to read cached manifest at %s", cb.ManifestPath)
		}
		cb.Manifest = m
	}

	return true, nil
}
