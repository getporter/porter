package cnabprovider

import (
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/valuesource"
)

// CNABProvider is the interface Porter uses to communicate with the CNAB runtime
type CNABProvider interface {
	LoadBundle(bundleFile string) (*bundle.Bundle, error)
	LoadParameterSets(paramSets []string) (valuesource.Set, error)
	Install(arguments ActionArguments) error
	Upgrade(arguments ActionArguments) error
	Invoke(action string, arguments ActionArguments) error
	Uninstall(arguments ActionArguments) error
}
