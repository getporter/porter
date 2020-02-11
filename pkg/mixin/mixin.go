package mixin

import (
	"get.porter.sh/porter/pkg/pkgmgmt"
)

func IsCoreMixinCommand(value string) bool {
	switch value {
	case "install", "upgrade", "uninstall", "build", "schema", "version":
		return true
	default:
		return false
	}
}

// MixinProvider manages Porter's mixins: installing, listing, upgrading and
// general communication.
type MixinProvider interface {
	pkgmgmt.PackageManager

	// GetSchema requests the manifest schema from the mixin.
	GetSchema(name string) (string, error)
}
