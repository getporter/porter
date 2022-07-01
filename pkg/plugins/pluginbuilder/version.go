package pluginbuilder

import (
	"context"
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/porter/version"
)

var (
	// Version of your plugin
	// This is set by ldflags when you compile the go binary for your plugin
	Version string

	// Commit hash of your plugin
	// This is set by ldflags when you compile the go binary for your plugin
	Commit string
)

// PrintVersion introspects the configured plugin and returns metadata about the
// plugin. This is the plugin's implementation for the porter plugins list
// command.
func (p *PorterPlugin) PrintVersion(ctx context.Context, opts version.Options) error {
	metadata := plugins.Metadata{
		Metadata: pkgmgmt.Metadata{
			Name: p.Name(),
			VersionInfo: pkgmgmt.VersionInfo{
				Version: Version,
				Commit:  Commit,
				Author:  p.opts.Author,
			},
		},
		Implementations: make([]plugins.Implementation, 0, len(p.opts.RegisteredPlugins)),
	}

	for key := range p.opts.RegisteredPlugins {
		parts := strings.Split(key, ".")
		if len(parts) != 3 {
			return fmt.Errorf("the plugin is configured with an invalid set of plugin implementations: plugin keys should have 3 parts but got %s", key)
		}
		pluginInterface := parts[0]
		pluginImplementation := parts[2]
		metadata.Implementations = append(metadata.Implementations, plugins.Implementation{
			Type: pluginInterface,
			Name: pluginImplementation,
		})
	}
	return version.PrintVersion(p.porterConfig.Context, opts, metadata)
}
