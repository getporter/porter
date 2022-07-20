package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"
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
	parentDir, err := fs.GetPackagesDir()
	if err != nil {
		return nil, fmt.Errorf("could not get package directory:%w", err)
	}

	files, err := fs.FileSystem.ReadDir(parentDir)
	if err != nil {
		return nil, fmt.Errorf("could not list the contents of the %s directory %q: %w", fs.PackageType, parentDir, err)
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

func (fs *FileSystem) GetMetadata(ctx context.Context, name string) (pkgmgmt.PackageMetadata, error) {
	ctx, span := tracing.StartSpan(ctx, attribute.String("package.type", fs.PackageType), attribute.String("package.name", name))
	defer span.EndSpan()

	pkgDir, err := fs.GetPackageDir(name)
	if err != nil {
		return nil, span.Error(err)
	}
	r := NewRunner(name, pkgDir, false)

	// Copy the existing context and tweak to pipe the output differently
	jsonB := &bytes.Buffer{}
	pkgContext := *fs.Context
	pkgContext.Out = jsonB
	if span.ShouldLog(zapcore.DebugLevel) {
		pkgContext.Err = ioutil.Discard
	}
	r.Context = &pkgContext

	cmd := pkgmgmt.CommandOptions{Command: "version --output json", PreRun: fs.PreRun}
	err = r.Run(ctx, cmd)
	if err != nil {
		return nil, span.Error(err)
	}

	result := fs.BuildMetadata()
	err = json.Unmarshal(jsonB.Bytes(), &result)
	if err != nil {
		return nil, span.Error(err)
	}

	return result, nil
}

func (fs *FileSystem) Run(ctx context.Context, pkgContext *portercontext.Context, name string, commandOpts pkgmgmt.CommandOptions) error {
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
	return r.Run(ctx, commandOpts)
}

func (fs *FileSystem) GetPackagesDir() (string, error) {
	home, err := fs.GetHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, fs.PackageType), nil
}

func (fs *FileSystem) GetPackageDir(name string) (string, error) {
	parentDir, err := fs.GetPackagesDir()
	if err != nil {
		return "", err
	}

	pkgDir := filepath.Join(parentDir, name)
	dirExists, err := fs.FileSystem.DirExists(pkgDir)
	if err != nil {
		return "", fmt.Errorf("%s %s not accessible at %s: %w", fs.PackageType, name, pkgDir, err)
	}
	if !dirExists {
		return "", fmt.Errorf("%s %s not installed in %s", fs.PackageType, name, pkgDir)
	}

	return pkgDir, nil
}

func (fs *FileSystem) BuildClientPath(pkgDir string, name string) string {
	return filepath.Join(pkgDir, name) + pkgmgmt.FileExt
}
