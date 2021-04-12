package mixin

import (
	"get.porter.sh/porter/pkg/pkgmgmt"
)

type InstallOptions struct {
	pkgmgmt.InstallOptions
}

func (o *InstallOptions) Validate(args []string) error {
	o.PackageType = "mixin"
	return o.InstallOptions.Validate(args)
}
