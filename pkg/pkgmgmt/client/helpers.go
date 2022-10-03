package client

import (
	"context"
	"fmt"
	"path"
	"testing"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
)

var _ pkgmgmt.PackageManager = &TestPackageManager{}

// TestPackageManager helps us test mixins/plugins in our unit tests without
// actually hitting any real executables on the file system.
type TestPackageManager struct {
	PkgType       string
	Packages      []pkgmgmt.PackageMetadata
	RunAssertions []func(pkgContext *portercontext.Context, name string, commandOpts pkgmgmt.CommandOptions) error

	// called keeps track of which mixins/plugins were called
	called map[string]int
}

// GetCalled tracks how many times each package was called
func (p *TestPackageManager) GetCalled() map[string]int {
	return p.called
}

func (p *TestPackageManager) recordCalled(name string) {
	if p.called == nil {
		p.called = make(map[string]int, 1)
	}

	p.called[name]++
}

func (p *TestPackageManager) List() ([]string, error) {
	names := make([]string, 0, len(p.Packages))
	for _, pkg := range p.Packages {
		names = append(names, pkg.GetName())
	}
	return names, nil
}

func (p *TestPackageManager) GetPackageDir(name string) (string, error) {
	return path.Join("/home/myuser/.porter", p.PkgType, name), nil
}

func (p *TestPackageManager) GetMetadata(ctx context.Context, name string) (pkgmgmt.PackageMetadata, error) {
	for _, pkg := range p.Packages {
		if pkg.GetName() == name {
			p.recordCalled(name)
			return pkg, nil
		}
	}
	return nil, fmt.Errorf("%s %s not installed", p.PkgType, name)
}

func (p *TestPackageManager) Install(ctx context.Context, opts pkgmgmt.InstallOptions) error {
	// do nothing
	return nil
}

func (p *TestPackageManager) Uninstall(ctx context.Context, opts pkgmgmt.UninstallOptions) error {
	// do nothing
	return nil
}

func (p *TestPackageManager) Run(ctx context.Context, pkgContext *portercontext.Context, name string, commandOpts pkgmgmt.CommandOptions) error {
	for _, assert := range p.RunAssertions {
		p.recordCalled(name)
		err := assert(pkgContext, name, commandOpts)
		if err != nil {
			return err
		}
	}
	return nil
}

type TestRunner struct {
	*Runner
	TestContext *portercontext.TestContext
}

// NewTestRunner initializes a test runner, with the output buffered, and an in-memory file system.
func NewTestRunner(t *testing.T, name string, pkgType string, runtime bool) *TestRunner {
	c := portercontext.NewTestContext(t)
	pkgDir := fmt.Sprintf("/home/myuser/.porter/%s/%s", pkgType, name)
	r := &TestRunner{
		Runner:      NewRunner(name, pkgDir, runtime),
		TestContext: c,
	}
	r.Context = c.Context

	// Setup Porter home
	c.FileSystem.Create("/home/myuser/.porter/porter")
	c.FileSystem.Create("/home/myuser/.porter/runtimes/porter-runtime")
	c.FileSystem.Create(path.Join(pkgDir, name))
	c.FileSystem.Create(path.Join(pkgDir, "runtimes", name+"-runtime"))

	return r
}
