package configadapter

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/porter/pkg/config"
	"github.com/pkg/errors"
)

type Stamp struct {
	ManifestDigest string `json:"manifestDigest"`
}

func (c *ManifestConverter) GenerateStamp() Stamp {
	stamp := Stamp{}

	digest, err := c.digestManifest()
	if err != nil {
		// The digest is only used to decide if we need to rebuild, it is not an error condition to not
		// have a digest.
		fmt.Fprintln(c.Err, errors.Wrap(err, "WARNING: Could not digest the porter manifest file"))
		stamp.ManifestDigest = "unknown"
	} else {
		stamp.ManifestDigest = digest
	}

	return stamp
}

func (c *ManifestConverter) digestManifest() (string, error) {
	if exists, _ := c.FileSystem.Exists(c.Manifest.ManifestPath); !exists {
		return "", errors.Errorf("the specified porter configuration file %s does not exist", c.Manifest.ManifestPath)
	}

	data, err := c.FileSystem.ReadFile(c.Manifest.ManifestPath)
	if err != nil {
		return "", errors.Wrapf(err, "could not read manifest at %q", c.Manifest.ManifestPath)
	}

	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func LoadStamp(bun *bundle.Bundle) (*Stamp, error) {
	data := bun.Custom[config.CustomBundleKey]

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal the porter stamp %q", string(dataB))
	}

	stamp := &Stamp{}
	err = json.Unmarshal(dataB, stamp)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal the porter stamp %q", string(dataB))
	}

	return stamp, nil
}
