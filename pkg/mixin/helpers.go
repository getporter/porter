package mixin

import (
	"fmt"
	"io/ioutil"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/client"
)

type TestMixinProvider struct {
	client.TestPackageManager
}

// NewTestMixinProvider helps us test Porter.Mixins in our unit tests without actually hitting any real plugins on the file system.
func NewTestMixinProvider() *TestMixinProvider {
	packages := []pkgmgmt.PackageMetadata{
		&Metadata{
			Name: "exec",
			VersionInfo: pkgmgmt.VersionInfo{
				Version: "v1.0",
				Commit:  "abc123",
				Author:  "Porter Authors",
			},
		},
	}

	provider := TestMixinProvider{
		TestPackageManager: client.TestPackageManager{
			PkgType:  "mixins",
			Packages: packages,
		},
	}

	provider.RunAssertions = []func(pkgContext *context.Context, name string, commandOpts pkgmgmt.CommandOptions){
		provider.PrintExecOutput,
	}

	return &provider
}

func (p *TestMixinProvider) PrintExecOutput(pkgContext *context.Context, name string, commandOpts pkgmgmt.CommandOptions) {
	if commandOpts.Command == "build" {
		fmt.Fprintln(pkgContext.Out, "# exec mixin has no buildtime dependencies")
	}
}

func (p *TestMixinProvider) GetSchema(name string) (string, error) {
	b, err := ioutil.ReadFile("../exec/schema/exec.json")
	return string(b), err
}
