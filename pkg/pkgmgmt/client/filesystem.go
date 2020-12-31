package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/pkg/errors"
)

var _ pkgmgmt.PackageManager = &FileSystem{}

type PackageMetadataBuilder func() pkgmgmt.PackageMetadata

func NewFileSystem(config *config.Config, pkgType string) *FileSystem {
	return &FileSystem{
		Config:        config,
		PackageType:   pkgType,
		BuildMetadata: func() pkgmgmt.PackageMetadata { return pkgmgmt.Metadata{} },
	}
}

type FileSystem struct {
	*config.Config

	// PackageType is the type of package managed by this instance of the
	// package manager. It must also correspond to the directory name in
	// PORTER_HOME.
	PackageType string

	// PreRun is executed before commands are run against a package, giving
	// consumers a chance to tweak the command first.
	PreRun pkgmgmt.PreRunHandler

	// BuildMetadata allows mixins/plugins to supply the proper struct that
	// represents its package metadata.
	BuildMetadata PackageMetadataBuilder
}

func (fs *FileSystem) List() ([]string, error) {
	parentDir := fs.GetPackagesDir()
	files, err := fs.FileSystem.ReadDir(parentDir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list the contents of the %s directory %q", fs.PackageType, parentDir)
	}

	names := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		names = append(names, file.Name())
	}

	return names, nil
}

func (fs *FileSystem) GetMetadata(name string) (pkgmgmt.PackageMetadata, error) {
	pkgDir, err := fs.GetPackageDir(name)
	if err != nil {
		return nil, err
	}
	r := NewRunner(name, pkgDir, false)

	// Copy the existing context and tweak to pipe the output differently
	jsonB := &bytes.Buffer{}
	var pkgContext context.Context
	pkgContext = *fs.Context
	pkgContext.Out = jsonB
	if !fs.Debug {
		pkgContext.Err = ioutil.Discard
	}
	r.Context = &pkgContext

	cmd := pkgmgmt.CommandOptions{Command: "version --output json", PreRun: fs.PreRun}
	err = r.Run(cmd)
	if err != nil {
		return nil, err
	}

	result := fs.BuildMetadata()
	err = json.Unmarshal(jsonB.Bytes(), &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (fs *FileSystem) Run(pkgContext *context.Context, name string, commandOpts pkgmgmt.CommandOptions) error {
	pkgDir, err := fs.GetPackageDir(name)
	if err != nil {
		return err
	}

	r := NewRunner(name, pkgDir, commandOpts.Runtime)
	r.Context = pkgContext

	err = r.Validate()
	if err != nil {
		return err
	}

	commandOpts.PreRun = fs.PreRun
	return r.Run(commandOpts)
}

func (fs *FileSystem) GetPackagesDir() string {
	return filepath.Join(fs.GetHomeDir(), fs.PackageType)
}

func (fs *FileSystem) GetPackageDir(name string) (string, error) {
	pkgDir := filepath.Join(fs.GetPackagesDir(), name)
	dirExists, err := fs.FileSystem.DirExists(pkgDir)
	if err != nil {
		return "", errors.Wrapf(err, "%s %s not accessible at %s", fs.PackageType, name, pkgDir)
	}
	if !dirExists {
		return "", fmt.Errorf("%s %s not installed in %s", fs.PackageType, name, pkgDir)
	}

	return pkgDir, nil
}

func (fs *FileSystem) BuildClientPath(pkgDir string, name string) string {
	return filepath.Join(pkgDir, name) + pkgmgmt.FileExt
}
