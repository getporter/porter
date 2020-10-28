package plugins

import (
	"get.porter.sh/porter/pkg/pkgmgmt"
)

// PluginProvider manages Porter's plugins: installing, listing, upgrading and
// general communication.
type PluginProvider interface {
	pkgmgmt.PackageManager
}
