package plugins

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/client"
	"github.com/gobuffalo/packr/v2"
)

const (
	Directory = "plugins"
)

var _ PluginProvider = &PackageManager{}

type PackageManager struct {
	*client.FileSystem
}

func NewPackageManager(c *config.Config) *PackageManager {
	client := PackageManager{
		FileSystem: client.NewFileSystem(c, Directory),
	}

	client.BuildMetadata = func() pkgmgmt.PackageMetadata {
		return &Metadata{}
	}

	return &client
}

var _ pkgmgmt.PackageMetadata = Metadata{}

// Metadata about an installed plugin.
type Metadata struct {
	pkgmgmt.Metadata `json:",inline" yaml:",inline"`
	Implementations  []Implementation `json:"implementations" yaml:"implementations"`
}

// Implementation stores implementation type (e.g. storage) and its name (e.g.
// s3, mongo)
type Implementation struct {
	Type string `json:"type" yaml:"type"`
	Name string `json:"implementation" yaml:"name"`
}

// GetDirectoryListings returns a directory/list of plugins available to install
func GetDirectoryListings() *packr.Box {
	// TODO: Decouple listing from CLI: https://github.com/deislabs/porter/issues/908
	return packr.New("get.porter.sh/porter/pkg/plugins/directory", "./directory")
}
