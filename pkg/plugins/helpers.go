package plugins

import (
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/client"
)

// NewTestPluginProvider helps us test Porter.Plugins in our unit tests without actually hitting any real plugins on the file system.
func NewTestPluginProvider() *client.TestPackageManager {
	v := pkgmgmt.VersionInfo{
		Version: "v1.0",
		Commit:  "abc123",
		Author:  "Porter Authors",
	}
	impl := []Implementation{
		{Type: "storage", Name: "blob"},
		{Type: "storage", Name: "mongo"},
	}
	return &client.TestPackageManager{
		PkgType: "plugins",
		Packages: []pkgmgmt.PackageMetadata{
			&Metadata{Metadata: pkgmgmt.Metadata{Name: "plugin1", VersionInfo: v}, Implementations: impl},
			&Metadata{Metadata: pkgmgmt.Metadata{Name: "plugin2", VersionInfo: v}, Implementations: impl},
			&Metadata{Metadata: pkgmgmt.Metadata{Name: "unknown", VersionInfo: v}},
		},
	}
}
