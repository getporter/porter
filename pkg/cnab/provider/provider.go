package cnabprovider

import (
	"context"

	"get.porter.sh/porter/pkg/storage"

	"get.porter.sh/porter/pkg/cnab"
)

// CNABProvider is the interface Porter uses to communicate with the CNAB runtime
type CNABProvider interface {
	LoadBundle(bundleFile string) (cnab.ExtendedBundle, error)
	Execute(ctx context.Context, arguments ActionArguments) (storage.Run, storage.Result, error)
}
