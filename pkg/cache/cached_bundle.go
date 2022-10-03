package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/cnabio/cnab-to-oci/relocation"
)

// CachedBundle represents a bundle pulled from a registry that has been cached to
// the filesystem.
type CachedBundle struct {
	// cacheDir is the cache directory for the bundle (not the general cache directory)
	cacheDir string

	// BundleReference contains common bundle metadata, such as the definition.
	cnab.BundleReference

	// BundlePath is the location of the bundle.json in the cache.
	BundlePath string

	// ManifestPath is the optional location of the porter.yaml in the cache.
	ManifestPath string

	// RelocationFilePath is the optional location of the relocation file in the cache.
	RelocationFilePath string
}

// GetBundleID is the unique ID of the cached bundle.
func (cb *CachedBundle) GetBundleID() string {
	// hash the tag, tags have characters that won't work as part of a path
	// so hashing here to get a path friendly name
	bid := md5.Sum([]byte(cb.Reference.String()))
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

// BuildMetadataPath generates the location of the cache metadata.
func (cb *CachedBundle) BuildMetadataPath() string {
	return filepath.Join(cb.cacheDir, "metadata.json")
}

// Load starts from the bundle tag, and hydrates the cached bundle from the cache.
func (cb *CachedBundle) Load(cxt *portercontext.Context) (bool, error) {
	// Check that the bundle exists
	cb.BundlePath = cb.BuildBundlePath()
	metaPath := cb.BuildMetadataPath()
	metaExists, err := cxt.FileSystem.Exists(metaPath)
	if err != nil {
		return false, fmt.Errorf("unable to access bundle metadata %s at %s: %w", cb.Reference, metaPath, err)
	}
	if !metaExists {
		// consider this a miss, recache with the metadata
		return false, nil
	}
	var meta Metadata
	err = encoding.UnmarshalFile(cxt.FileSystem, metaPath, &meta)
	if err != nil {
		return false, fmt.Errorf("unable to parse cached bundle metadata %s at %s: %w", cb.Reference, metaPath, err)
	}
	cb.Digest = meta.Digest

	// Check for the optional relocation mapping next to it
	reloPath := cb.BuildRelocationFilePath()
	reloExists, err := cxt.FileSystem.Exists(reloPath)
	if err != nil {
		return true, fmt.Errorf("unable to read relocation mapping %s at %s: %w", cb.Reference, reloPath, err)
	}
	if reloExists {
		cb.RelocationFilePath = reloPath
	}

	// Check for the optional manifest
	manifestPath := cb.BuildManifestPath()
	manifestExists, err := cxt.FileSystem.Exists(manifestPath)
	if err != nil {
		return true, fmt.Errorf("unable to read manifest %s at %s: %w", cb.Reference, manifestPath, err)
	}
	if manifestExists {
		cb.ManifestPath = manifestPath
	}

	bun, err := cnab.LoadBundle(cxt, cb.BundlePath)
	if err != nil {
		return true, fmt.Errorf("unable to parse cached bundle file at %s: %w", cb.BundlePath, err)
	}
	cb.Definition = bun

	if cb.RelocationFilePath != "" {
		data, err := cxt.FileSystem.ReadFile(cb.RelocationFilePath)
		if err != nil {
			return true, fmt.Errorf("unable to read cached relocation file at %s: %w", cb.RelocationFilePath, err)
		}

		reloMap := relocation.ImageRelocationMap{}
		err = json.Unmarshal(data, &reloMap)
		if err != nil {
			return true, fmt.Errorf("unable to parse cached relocation file at %s: %w", cb.RelocationFilePath, err)
		}
		cb.RelocationMap = reloMap
	}

	return true, nil
}
