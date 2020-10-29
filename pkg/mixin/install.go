package mixin

import (
	"get.porter.sh/porter/pkg/pkgmgmt"
)

const DefaultFeedUrl = "https://cdn.porter.sh/mixins/atom.xml"

type InstallOptions struct {
	pkgmgmt.InstallOptions
}

func (o *InstallOptions) Validate(args []string) error {
	o.DefaultFeedURL = DefaultFeedUrl
	return o.InstallOptions.Validate(args)
}
