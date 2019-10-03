package construction

import (
	"os"
	"path/filepath"

	"github.com/deislabs/cnab-go/imagestore"
	"github.com/deislabs/cnab-go/imagestore/ocilayout"
	"github.com/deislabs/cnab-go/imagestore/remote"
)

// NewConstructor creates an image store constructor which will, if necessary, create archive contents.
func NewConstructor(remoteRepos bool) (imagestore.Constructor, error) {
	// infer the concrete type of the image store from the input parameters
	if remoteRepos {
		return remote.Create, nil
	}
	return ocilayout.Create, nil
}

// NewLocatingConstructor creates an image store constructor which will, if necessary, find existing archive contents.
func NewLocatingConstructor() imagestore.Constructor {
	return func(options ...imagestore.Option) (imagestore.Store, error) {
		parms := imagestore.Create(options...)
		if thin(parms.ArchiveDir) {
			return remote.Create()
		}
		return ocilayout.LocateOciLayout(parms.ArchiveDir)
	}
}

func thin(archiveDir string) bool {
	// If there is no archive directory, the bundle is thin
	if archiveDir == "" {
		return true
	}

	// If there is an archive directory, the bundle is thin if and only if the archive directory has no artifacts/
	// subdirectory
	layoutDir := filepath.Join(archiveDir, "artifacts")
	_, err := os.Stat(layoutDir)
	return os.IsNotExist(err)
}
