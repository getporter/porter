package mixin

import (
	"context"

	"get.porter.sh/porter/pkg/pkgmgmt"
)

func IsCoreMixinCommand(value string) bool {
	switch value {
	case "install", "upgrade", "uninstall", "build", "lint", "schema", "version":
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
	GetSchema(ctx context.Context, name string) (string, error)
}
