package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/pkg/errors"
)

type BundleCache interface {
	FindBundle(tag string) (bun CachedBundle, found bool, err error)
	StoreBundle(tag string, bun bundle.Bundle, reloMap *relocation.ImageRelocationMap) (CachedBundle, error)
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
func (c *Cache) FindBundle(bundleTag string) (CachedBundle, bool, error) {
	cb := CachedBundle{
		Tag: bundleTag,
	}

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
func (c *Cache) StoreBundle(bundleTag string, bun bundle.Bundle, reloMap *relocation.ImageRelocationMap) (CachedBundle, error) {
	cb := CachedBundle{
		Tag:           bundleTag,
		Bundle:        bun,
		RelocationMap: reloMap,
	}

	cacheDir, err := c.GetCacheDir()
	if err != nil {
		return CachedBundle{}, err
	}
	cb.SetCacheDir(cacheDir)

	cb.BundlePath = cb.BuildBundlePath()
	err = c.FileSystem.MkdirAll(filepath.Dir(cb.BundlePath), os.ModePerm)
	if err != nil {
		return CachedBundle{}, errors.Wrap(err, "unable to create cache directory")
	}

	f, err := c.FileSystem.OpenFile(cb.BundlePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return CachedBundle{}, errors.Wrapf(err, "error creating cnab/bundle.json for %s", bundleTag)
	}

	_, err = bun.WriteTo(f)
	if err != nil {
		return CachedBundle{}, errors.Wrapf(err, "error writing to cnab/bundle.json for %s", bundleTag)
	}

	err = c.cacheManifest(&cb)
	if err != nil {
		return CachedBundle{}, err
	}

	// we wrote the bundle, now lets store a relocation mapping in cnab/ and return the path
	if reloMap != nil && len(*reloMap) > 0 {
		cb.RelocationFilePath = cb.BuildRelocationFilePath()
		f, err = c.FileSystem.OpenFile(cb.RelocationFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		defer f.Close()
		if err != nil {
			return CachedBundle{}, errors.Wrapf(err, "error creating cnab/relocation-mapping.json for %s", bundleTag)
		}

		b, err := json.Marshal(reloMap)
		if err != nil {
			return CachedBundle{}, errors.Wrapf(err, "couldn't marshall relocation mapping for %s", bundleTag)
		}

		_, err = f.Write(b)
		if err != nil {
			return CachedBundle{}, errors.Wrapf(err, "couldn't write relocation mapping for %s", bundleTag)
		}

	}

	return cb, nil
}

// cacheManifest extracts the porter.yaml from the bundle, if present and caches it
// in the same cache directory as the rest of the bundle.
func (c *Cache) cacheManifest(cb *CachedBundle) error {
	if configadapter.IsPorterBundle(cb.Bundle) {
		stamp, err := configadapter.LoadStamp(cb.Bundle)
		if err != nil {
			fmt.Fprintf(c.Err, "WARNING: Bundle %s was created by porter but could not find a porter manifest embedded. This may be because it was created by an older version of Porter.\n", cb.Tag)
			return nil
		}

		cb.ManifestPath = cb.BuildManifestPath()
		err = stamp.WriteManifest(c.Context, cb.ManifestPath)
		if err != nil {
			return errors.Wrapf(err, "error writing porter.yaml for %s", cb.Tag)
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
