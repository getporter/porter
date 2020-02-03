package client

import (
	"fmt"
	"path"
	"testing"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/pkgmgmt"
)

var _ pkgmgmt.PackageManager = &TestPackageManager{}

// TestPackageManager helps us test mixins/plugins in our unit tests without
// actually hitting any real executables on the file system.
type TestPackageManager struct {
	PkgType       string
	Packages      []pkgmgmt.PackageMetadata
	RunAssertions []func(pkgContext *context.Context, name string, commandOpts pkgmgmt.CommandOptions)
}

func (p *TestPackageManager) List() ([]string, error) {
	names := make([]string, 0, len(p.Packages))
	for _, pkg := range p.Packages {
		names = append(names, pkg.GetName())
	}
	return names, nil
}

func (p *TestPackageManager) GetPackageDir(name string) (string, error) {
	return path.Join("/root/.porter", p.PkgType, name), nil
}

func (p *TestPackageManager) GetMetadata(name string) (pkgmgmt.PackageMetadata, error) {
	for _, pkg := range p.Packages {
		if pkg.GetName() == name {
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("%s %s not installed", p.PkgType, name)
}

func (p *TestPackageManager) Install(o pkgmgmt.InstallOptions) error {
	// do nothing
	return nil
}

func (p *TestPackageManager) Uninstall(o pkgmgmt.UninstallOptions) error {
	// do nothing
	return nil
}

func (p *TestPackageManager) Run(pkgContext *context.Context, name string, commandOpts pkgmgmt.CommandOptions) error {
	for _, assert := range p.RunAssertions {
		assert(pkgContext, name, commandOpts)
	}
	return nil
}

type TestRunner struct {
	*Runner
	TestContext *context.TestContext
}

// NewTestRunner initializes a test runner, with the output buffered, and an in-memory file system.
func NewTestRunner(t *testing.T, name string, pkgType string, runtime bool) *TestRunner {
	c := context.NewTestContext(t)
	pkgDir := fmt.Sprintf("/root/.porter/%s/%s", pkgType, name)
	r := &TestRunner{
		Runner:      NewRunner(name, pkgDir, runtime),
		TestContext: c,
	}
	r.Context = c.Context

	// Setup Porter home
	c.FileSystem.Create("/root/.porter/porter")
	c.FileSystem.Create("/root/.porter/porter-runtime")
	c.FileSystem.Create(path.Join(pkgDir, name))
	c.FileSystem.Create(path.Join(pkgDir, name+"-runtime"))

	return r
}
