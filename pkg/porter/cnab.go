package porter

import cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"

// CNABProvider
type CNABProvider interface {
	Install(arguments cnabprovider.InstallArguments) error
	//Upgrade() error
	//Uninstall() error
}
