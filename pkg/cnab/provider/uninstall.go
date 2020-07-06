package cnabprovider

import (
	"github.com/cnabio/cnab-go/claim"
)

func (r *Runtime) Uninstall(args ActionArguments) error {
	return r.ExecuteAction(claim.ActionUninstall, args)
}
