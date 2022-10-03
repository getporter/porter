package mixin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/client"
	"get.porter.sh/porter/pkg/portercontext"
)

type TestMixinProvider struct {
	client.TestPackageManager

	// LintResults allows you to provide linter.Results for your unit tests.
	// It isn't of type linter.Results directly to avoid package cycles
	LintResults interface{}

	// ReturnBuildError will force the TestMixinProvider to return a build error
	// if set to true
	ReturnBuildError bool
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
		&Metadata{
			Name: "testmixin",
			VersionInfo: pkgmgmt.VersionInfo{
				Version: "v0.1.0",
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

	provider.RunAssertions = []func(pkgContext *portercontext.Context, name string, commandOpts pkgmgmt.CommandOptions) error{
		provider.PrintExecOutput,
	}

	return &provider
}

func (p *TestMixinProvider) PrintExecOutput(pkgContext *portercontext.Context, name string, commandOpts pkgmgmt.CommandOptions) error {
	switch commandOpts.Command {
	case "build":
		if p.ReturnBuildError {
			return errors.New("encountered build error")
		}
		fmt.Fprintln(pkgContext.Out, "# exec mixin has no buildtime dependencies")
	case "lint":
		b, _ := json.Marshal(p.LintResults)
		fmt.Fprintln(pkgContext.Out, string(b))
	}
	return nil
}

func (p *TestMixinProvider) GetSchema(ctx context.Context, name string) (string, error) {
	var schemaFile string
	switch name {
	case "exec":
		schemaFile = "../exec/schema/exec.json"
	case "testmixin":
		schemaFile = "../../cmd/testmixin/schema.json"
	default:
		return "", nil
	}
	b, err := ioutil.ReadFile(schemaFile)
	return string(b), err
}
