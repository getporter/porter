package cache

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/config"
	"github.com/pkg/errors"
)

type BundleCache interface {
	FindBundle(string) (string, bool, error)
	StoreBundle(string, *bundle.Bundle) (string, error)
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
// empty string and the boolean false value are returned.
func (c *cache) FindBundle(bundleTag string) (string, bool, error) {
	bid := getBundleID(bundleTag)
	bundleCnabDir, err := c.getCachedBundleCNABDir(bid)
	cachedBundlePath := filepath.Join(bundleCnabDir, "bundle.json")
	bExists, err := c.FileSystem.Exists(cachedBundlePath)
	if err != nil {
		return "", false, errors.Wrapf(err, "unable to read bundle %s at %s", bundleTag, cachedBundlePath)
	}
	if !bExists {
		return "", false, nil
	}
	return cachedBundlePath, true, nil

}

// StoreBundle will write a given bundle to the bundle cache, in a location derived
// from the bundleTag.
func (c *cache) StoreBundle(bundleTag string, bun *bundle.Bundle) (string, error) {
	bid := getBundleID(bundleTag)
	bundleCnabDir, err := c.getCachedBundleCNABDir(bid)
	cachedBundlePath := filepath.Join(bundleCnabDir, "bundle.json")
	err = c.FileSystem.MkdirAll(bundleCnabDir, os.ModePerm)
	if err != nil {
		return "", errors.Wrap(err, "unable to create cache directory")
	}
	f, err := c.FileSystem.OpenFile(cachedBundlePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return "", errors.Wrapf(err, "error creating cnab/bundle.json for %s", bundleTag)
	}
	_, err = bun.WriteTo(f)
	if err != nil {
		return "", errors.Wrapf(err, "error writing to cnab/bundle.json for %s", bundleTag)
	}
	return cachedBundlePath, nil
}

func (c *cache) GetCacheDir() (string, error) {
	home, err := c.GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "cache"), nil
}

func (c *cache) getCachedBundleCNABDir(bid string) (string, error) {
	cacheDir, err := c.GetCacheDir()
	if err != nil {
		return "", err
	}
	bundleDir := filepath.Join(cacheDir, string(bid))
	bundleCNABPath := filepath.Join(bundleDir, "cnab")
	return bundleCNABPath, nil

}

func getBundleID(bundleTag string) string {
	// hash the tag, tags have characters that won't work as part of a path
	// so hashing here to get a path friendly name
	bid := md5.Sum([]byte(bundleTag))
	return hex.EncodeToString(bid[:])
}
