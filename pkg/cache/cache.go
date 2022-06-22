package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"github.com/opencontainers/go-digest"
)

type BundleCache interface {
	FindBundle(tag cnab.OCIReference) (bun CachedBundle, found bool, err error)
	StoreBundle(bundleRef cnab.BundleReference) (CachedBundle, error)
	GetCacheDir() (string, error)
}

var _ BundleCache = &Cache{}

type Cache struct {
	*config.Config
}

func New(cfg *config.Config) BundleCache {
	return &Cache{
		Config: cfg,
	}
}

// FindBundle looks for a given bundle tag in the Porter bundle cache and
// returns the path to the bundle if it exists. If it is not found, an
// empty string and the boolean false value are returned. If the bundle is found,
// and a relocation mapping file is present, it will be returned as well. If the relocation
// is not found, an empty string is returned.
func (c *Cache) FindBundle(ref cnab.OCIReference) (CachedBundle, bool, error) {
	cb := CachedBundle{}
	cb.Reference = ref

	cacheDir, err := c.GetCacheDir()
	if err != nil {
		return CachedBundle{}, false, err
	}
	cb.SetCacheDir(cacheDir)

	found, err := cb.Load(c.Context)
	if err != nil {
		return CachedBundle{}, false, err
	}
	if !found {
		return CachedBundle{}, false, nil
	}
	return cb, true, nil

}

// StoreBundle will write a given bundle to the bundle cache, in a location derived
// from the bundleTag. If a relocation mapping is provided, it will be stored along side
// the bundle. If successful, returns the path to the bundle, along with the path to a
// relocation mapping, if provided. Otherwise, returns an error.
func (c *Cache) StoreBundle(bundleRef cnab.BundleReference) (CachedBundle, error) {
	cb := CachedBundle{BundleReference: bundleRef}

	cacheDir, err := c.GetCacheDir()
	if err != nil {
		return CachedBundle{}, err
	}
	cb.SetCacheDir(cacheDir)

	// Remove any previously cached bundle files
	err = c.FileSystem.RemoveAll(cb.cacheDir)
	if err != nil {
		return CachedBundle{}, fmt.Errorf("cannot remove existing cache directory %s: %w", cb.BundlePath, err)
	}

	cb.BundlePath = cb.BuildBundlePath()
	err = c.FileSystem.MkdirAll(filepath.Dir(cb.BundlePath), pkg.FileModeDirectory)
	if err != nil {
		return CachedBundle{}, fmt.Errorf("unable to create cache directory: %w", err)
	}

	f, err := c.FileSystem.OpenFile(cb.BundlePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, pkg.FileModeWritable)
	if err != nil {
		return CachedBundle{}, fmt.Errorf("error creating cnab/bundle.json for %s: %w", cb.Reference, err)
	}
	defer f.Close()

	_, err = cb.Definition.WriteTo(f)
	if err != nil {
		return CachedBundle{}, fmt.Errorf("error writing to cnab/bundle.json for %s: %w", cb.Reference, err)
	}

	err = c.cacheMetadata(&cb)
	if err != nil {
		return CachedBundle{}, err
	}

	err = c.cacheManifest(&cb)
	if err != nil {
		return CachedBundle{}, err
	}

	// we wrote the bundle, now lets store a relocation mapping in cnab/ and return the path
	if len(cb.RelocationMap) > 0 {
		cb.RelocationFilePath = cb.BuildRelocationFilePath()
		f, err = c.FileSystem.OpenFile(cb.RelocationFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, pkg.FileModeWritable)
		if err != nil {
			return CachedBundle{}, fmt.Errorf("error creating cnab/relocation-mapping.json for %s: %w", cb.Reference, err)
		}
		defer f.Close()

		b, err := json.Marshal(cb.RelocationMap)
		if err != nil {
			return CachedBundle{}, fmt.Errorf("couldn't marshall relocation mapping for %s: %w", cb.Reference, err)
		}

		_, err = f.Write(b)
		if err != nil {
			return CachedBundle{}, fmt.Errorf("couldn't write relocation mapping for %s: %w", cb.Reference, err)
		}

	}

	return cb, nil
}

// cacheMetadata stores additional metadata about the bundle.
func (c *Cache) cacheMetadata(cb *CachedBundle) error {
	meta := Metadata{
		Reference: cb.Reference,
		Digest:    cb.Digest,
	}
	path := cb.BuildMetadataPath()
	return encoding.MarshalFile(c.FileSystem, path, meta)
}

// Metadata associated with a cached bundle.
type Metadata struct {
	Reference cnab.OCIReference `json:"reference"`
	Digest    digest.Digest     `json:"digest"`
}

// cacheManifest extracts the porter.yaml from the bundle, if present and caches it
// in the same cache directory as the rest of the bundle.
func (c *Cache) cacheManifest(cb *CachedBundle) error {
	if cb.Definition.IsPorterBundle() {
		stamp, err := configadapter.LoadStamp(cb.Definition)
		if err != nil {
			fmt.Fprintf(c.Err, "WARNING: Bundle %s was created by porter but could not load the Porter stamp. This may be because it was created by an older version of Porter.\n", cb.Reference)
			return nil
		}

		if stamp.EncodedManifest == "" {
			fmt.Fprintf(c.Err, "WARNING: Bundle %s was created by porter but could not find a porter manifest embedded. This may be because it was created by an older version of Porter.\n", cb.Reference)
			return nil
		}

		cb.ManifestPath = cb.BuildManifestPath()
		err = stamp.WriteManifest(c.Context, cb.ManifestPath)
		if err != nil {
			return fmt.Errorf("error writing porter.yaml for %s: %w", cb.Reference, err)
		}
	}

	return nil
}

func (c *Cache) GetCacheDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "cache"), nil
}
