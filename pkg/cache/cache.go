package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/docker/cnab-to-oci/relocation"
	"github.com/pkg/errors"
)

type BundleCache interface {
	FindBundle(string) (string, string, bool, error)
	StoreBundle(string, *bundle.Bundle, relocation.ImageRelocationMap) (string, string, error)
	GetCacheDir() (string, error)
}

type cache struct {
	*config.Config
}

func New(cfg *config.Config) BundleCache {
	return &cache{
		Config: cfg,
	}
}

// FindBundle looks for a given bundle tag in the Porter bundle cache and
// returns the path to the bundle if it exists. If it is not found, an
// empty string and the boolean false value are returned. If the bundle is found,
// and a relocation mapping file is present, it will be returned as well. If the relocation
// is not found, an empty string is returned.
func (c *cache) FindBundle(bundleTag string) (string, string, bool, error) {
	bid := getBundleID(bundleTag)
	bundleCnabDir, err := c.getBundleCacheDir(bid)
	if err != nil {
		return "", "", false, err
	}

	cachedBundlePath := filepath.Join(bundleCnabDir, "bundle.json")
	cachedReloPath := filepath.Join(bundleCnabDir, "relocation-mapping.json")
	bExists, err := c.FileSystem.Exists(cachedBundlePath)
	if err != nil {
		return "", "", false, errors.Wrapf(err, "unable to read bundle %s at %s", bundleTag, cachedBundlePath)
	}
	if !bExists {
		return "", "", false, nil
	}
	//check for a relocation mapping next to it
	rExists, err := c.FileSystem.Exists(cachedReloPath)
	if err != nil {
		return "", "", false, errors.Wrapf(err, "unable to read relocation mapping %s at %s", bundleTag, cachedReloPath)
	}
	if rExists {
		return cachedBundlePath, cachedReloPath, true, nil
	}
	return cachedBundlePath, "", true, nil

}

// StoreBundle will write a given bundle to the bundle cache, in a location derived
// from the bundleTag. If a relocation mapping is provided, it will be stored along side
// the bundle. If successful, returns the path to the bundle, along with the path to a
// relocation mapping, if provided. Otherwise, returns an error.
func (c *cache) StoreBundle(bundleTag string, bun *bundle.Bundle, reloMap relocation.ImageRelocationMap) (string, string, error) {
	bid := getBundleID(bundleTag)
	bundleCacheDir, err := c.getBundleCacheDir(bid)
	if err != nil {
		return "", "", err
	}

	bundleCnabDir := filepath.Join(bundleCacheDir, "cnab")
	cachedBundlePath := filepath.Join(bundleCnabDir, "bundle.json")
	err = c.FileSystem.MkdirAll(bundleCnabDir, os.ModePerm)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to create cache directory")
	}
	f, err := c.FileSystem.OpenFile(cachedBundlePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return "", "", errors.Wrapf(err, "error creating cnab/bundle.json for %s", bundleTag)
	}
	_, err = bun.WriteTo(f)
	if err != nil {
		return "", "", errors.Wrapf(err, "error writing to cnab/bundle.json for %s", bundleTag)
	}

	err = c.cacheManifest(bundleTag, bun, bundleCacheDir)
	if err != nil {
		return "", "", err
	}

	// we wrote the bundle, now lets store a relocation mapping in cnab/ and return the path
	if len(reloMap) > 0 {
		cachedReloPath := filepath.Join(bundleCnabDir, "relocation-mapping.json")
		f, err = c.FileSystem.OpenFile(cachedReloPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		defer f.Close()
		if err != nil {
			return "", "", errors.Wrapf(err, "error creating cnab/relocation-mapping.json for %s", bundleTag)
		}
		b, err := json.Marshal(reloMap)
		if err != nil {
			return "", "", errors.Wrapf(err, "couldn't marshall relocation mapping for %s", bundleTag)
		}
		_, err = f.Write(b)
		if err != nil {
			return "", "", errors.Wrapf(err, "couldn't write relocation mapping for %s", bundleTag)
		}

		return cachedBundlePath, cachedReloPath, nil
	}

	return cachedBundlePath, "", nil
}

// cacheManifest extracts the porter.yaml from the bundle, if present and caches it
// in the same cache directory as the rest of the bundle.
func (c *cache) cacheManifest(bundleTag string, bun *bundle.Bundle, cacheDir string) error {
	if configadapter.IsPorterBundle(bun) {
		stamp, err := configadapter.LoadStamp(bun)
		if err != nil {
			fmt.Fprintf(c.Err, "WARNING: Bundle %s was created by porter but could not find a porter manifest embedded. This may be because it was create by an older version of Porter.\n", bundleTag)
			return nil
		}

		cachedManifestPath := filepath.Join(cacheDir, config.Name)
		err = stamp.WriteManifest(c.Context, cachedManifestPath)
		if err != nil {
			return errors.Wrapf(err, "error writing to %s for %s", cachedManifestPath, bundleTag)
		}
	}

	return nil
}

func (c *cache) GetCacheDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "cache"), nil
}

func (c *cache) getBundleCacheDir(bundleID string) (string, error) {
	cacheDir, err := c.GetCacheDir()
	if err != nil {
		return "", err
	}
	bundleCacheDir := filepath.Join(cacheDir, bundleID)
	return bundleCacheDir, nil
}

func getBundleID(bundleTag string) string {
	// hash the tag, tags have characters that won't work as part of a path
	// so hashing here to get a path friendly name
	bid := md5.Sum([]byte(bundleTag))
	return hex.EncodeToString(bid[:])
}
