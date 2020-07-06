package cnabprovider

import "github.com/cnabio/cnab-go/claim"

func (r *Runtime) Upgrade(args ActionArguments) error {
	return r.ExecuteAction(claim.ActionUpgrade, args)
}
