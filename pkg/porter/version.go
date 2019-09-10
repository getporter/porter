package porter

import (
	"github.com/deislabs/porter/pkg"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/porter/version"
)

func (p *Porter) PrintVersion(opts version.Options) error {
	metadata := mixin.Metadata{
		Name: "porter",
		VersionInfo: mixin.VersionInfo{
			Version: pkg.Version,
			Commit:  pkg.Commit,
		},
	}
	return version.PrintVersion(p.Context, opts, metadata)
}
