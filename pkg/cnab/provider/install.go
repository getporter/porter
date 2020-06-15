package cnabprovider

import (
	"github.com/cnabio/cnab-go/claim"
)

func (r *Runtime) Install(args ActionArguments) error {
	return r.ExecuteAction(claim.ActionInstall, args)
}
