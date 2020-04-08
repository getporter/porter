package cnabprovider

import "github.com/cnabio/cnab-go/bundle"

// CNABProvider is the interface Porter uses to communicate with the CNAB runtime
type CNABProvider interface {
	LoadBundle(bundleFile string) (*bundle.Bundle, error)
	Install(arguments ActionArguments) error
	Upgrade(arguments ActionArguments) error
	Invoke(action string, arguments ActionArguments) error
	Uninstall(arguments ActionArguments) error
}
