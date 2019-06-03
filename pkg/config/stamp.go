package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/pkg/errors"
)

type Stamp struct {
	ManifestDigest string `json:"manifestDigest"`
}

func (c *Config) GenerateStamp(m *Manifest) Stamp {
	stamp := Stamp{}

	digest, err := c.digestManifest(m)
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

func (c *Config) digestManifest(m *Manifest) (string, error) {
	if exists, _ := c.FileSystem.Exists(m.path); !exists {
		return "", errors.Errorf("the specified porter configuration file %s does not exist", m.path)
	}

	data, err := c.FileSystem.ReadFile(m.path)
	if err != nil {
		return "", errors.Wrapf(err, "could not read manifest at %q", m.path)
	}

	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func (c *Config) LoadStamp(bun bundle.Bundle) (*Stamp, error) {
	data := bun.Custom[CustomBundleKey]

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
