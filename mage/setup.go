package mage

import (
	"fmt"
	"os"

	"get.porter.sh/porter/mage/tools"
	"get.porter.sh/porter/pkg"
	"github.com/carolynvs/magex/pkg/gopath"
	"github.com/pkg/errors"
)

// ConfigureAgent sets up an Azure DevOps agent with EnsureMage and ensures
// that GOPATH/bin is in PATH.
func ConfigureAgent() error {
	err := tools.EnsureMage()
	if err != nil {
		return err
	}

	// Instruct Azure DevOps to add GOPATH/bin to PATH
	gobin := gopath.GetGopathBin()
	err = os.MkdirAll(gobin, pkg.FileModeDirectory)
	if err != nil {
		return errors.Wrapf(err, "could not mkdir -p %s", gobin)
	}
	fmt.Printf("Adding %s to the PATH\n", gobin)
	fmt.Printf("##vso[task.prependpath]%s\n", gobin)
	return nil
}
