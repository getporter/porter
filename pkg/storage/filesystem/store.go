package filesystem

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
)

var _ crud.Store = &Store{}

// Store is a local filesystem store that stores data in the porter home directory.
type Store struct {
	crud.Store
	config.Config
	logger hclog.Logger
}

func NewStore(c config.Config, l hclog.Logger) crud.Store {
	// Wrapping ourselves in a backing store so that our Connect is used.
	return crud.NewBackingStore(&Store{
		Config: c,
		logger: l,
	})
}

func (s *Store) Connect() error {
	if s.Store != nil {
		return nil
	}

	home, err := s.Config.GetHomeDir()
	if err != nil {
		return errors.Wrap(err, "could not determine home directory for filesystem storage")
	}

	s.logger.Info("PORTER HOME: " + home)

	s.Store = crud.NewFileSystemStore(home, NewFileExtensions())
	return s.CheckFilePermissions()
}

// CheckFilePermissions for files in PORTER_HOME and
// return an error if they are less restrictive than recommended.
func (s *Store) CheckFilePermissions() error {
	// File permissions on windows aren't returned by stat, and by default files
	// aren't shared with other users.
	if runtime.GOOS == "windows" {
		return nil
	}

	home, _ := s.GetHomeDir()
	s.logger.Info(fmt.Sprintf("Checking file permissions in %s\n", home))

	// Warn if any files that may contain sensitive information have the wrong permissions
	wantMode := os.FileMode(0600)

	checkFile := func(f string) error {
		info, err := s.FileSystem.Stat(f)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return errors.Wrapf(err, "could not check file permissions on %s", f)
		}

		perm := info.Mode().Perm()
		if wantMode != perm|wantMode {
			// %o prints the octal form of the permission bits, e.g. 755.
			return errors.Errorf("incorrect file permissions on %s (%o), it should be %o. Correct it manually or by running porter storage fix-permissions.", f, perm, wantMode)
		}

		return nil
	}

	checkDir := func(dir string) error {
		return s.FileSystem.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return err
			}

			if info.IsDir() {
				return nil
			}

			return checkFile(path)
		})
	}

	if s.ConfigFilePath != "" {
		if err := checkFile(s.ConfigFilePath); err != nil {
			return err
		}
	}

	dirs := []string{"claims", "outputs", "credentials", "parameters"}
	for _, dir := range dirs {
		if err := checkDir(filepath.Join(home, dir)); err != nil {
			return err
		}
	}
	return nil
}

func NewFileExtensions() map[string]string {
	ext := claim.NewClaimStoreFileExtensions()

	jsonExt := ".json"
	ext[credentials.ItemType] = jsonExt

	// TODO (carolynvs): change to parameters.ItemType once parameters move to cnab-go
	ext["parameters"] = jsonExt

	// Handle top level files, like schema.json
	ext[""] = jsonExt

	return ext
}
