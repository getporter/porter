package credentials

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/context"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// CredentialsMigration checks for legacy credentials in yaml format and makes a copy of them
// in json format next to them. This helps with porter switching formats without forcing a hard
// switch and making people do the migration themselves.
type credentialsMigration struct {
	*context.Context
	migrated []string
}

func NewCredentialsMigration(c *context.Context) *credentialsMigration {
	return &credentialsMigration{Context: c}
}

// Migrate accepts a path to a credential set directory and migrates every yaml file to json that doesn't already
// have a matching destination file (NAME.json).
func (m *credentialsMigration) Migrate(credsDir string) error {
	if hasCreds, _ := m.FileSystem.Exists(credsDir); !hasCreds {
		return nil
	}

	credFiles, err := m.FileSystem.ReadDir(credsDir)
	if err != nil {
		return err
	}

	lookup := make(map[string]struct{}, len(credFiles))
	for _, f := range credFiles {
		lookup[f.Name()] = struct{}{}
	}

	var convertErrs error
	for name := range lookup {
		if filepath.Ext(name) == ".yaml" {
			migratedName := strings.TrimSuffix(name, ".yaml") + ".json"
			if _, isMigrated := lookup[migratedName]; !isMigrated {
				legacyCredFile := filepath.Join(credsDir, name)
				err = m.ConvertToJson(legacyCredFile)
				if err != nil {
					convertErrs = multierror.Append(convertErrs, err)
				}
				lookup[migratedName] = struct{}{}
			}
		}
	}

	return convertErrs
}

// ConvertToJson accepts a path to a credential set formatted as yaml, and migrates it to json.
func (m *credentialsMigration) ConvertToJson(path string) error {
	if m.Debug {
		credName := filepath.Base(path)
		fmt.Fprintf(m.Err, "Converting credential %s from yaml to json. The old file is left next to it so you can remove it when you are sure you don't need it anymore.", credName)
	}

	b, err := m.FileSystem.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "could not read credentials at %s", path)
	}

	var cs credentials.CredentialSet
	err = yaml.Unmarshal(b, &cs)
	if err != nil {
		return errors.Wrapf(err, "could not parse credentials at %s", path)
	}

	fi, err := m.FileSystem.Stat(path)
	if err != nil {
		// Oh well, we tried, stamp them and move on
		cs.Modified = time.Now()
		cs.Created = cs.Modified
	} else {
		cs.Modified = fi.ModTime()
		cs.Created = cs.Modified // Created isn't available on all OS's so just set them the same
	}

	b, err = json.Marshal(cs)
	if err != nil {
		return errors.Wrapf(err, "could not convert credentials at %s to json", path)
	}

	// Change the filename from *.yaml to *.json
	destDir, destName := filepath.Split(path)
	destName = strings.TrimSuffix(destName, ".yaml") + ".json"
	dest := filepath.Join(destDir, destName)

	err = m.FileSystem.WriteFile(dest, b, 0644)
	if err != nil {
		errors.Wrapf(err, "could not save migrated credentials to %s", dest)
	}
	m.migrated = append(m.migrated, path)
	return nil
}
