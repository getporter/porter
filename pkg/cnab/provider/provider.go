package cnabprovider

import (
	"github.com/cnabio/cnab-go/bundle"
)

// CNABProvider is the interface Porter uses to communicate with the CNAB runtime
type CNABProvider interface {
	LoadBundle(bundleFile string) (bundle.Bundle, error)
	Execute(arguments ActionArguments) error
}
