package porter

import (
	"errors"

	"github.com/deislabs/porter/pkg/instance-storage/claimstore"
	"github.com/deislabs/porter/pkg/instance-storage/filesystem"
	"github.com/hashicorp/go-plugin"

	"github.com/deislabs/porter/pkg/plugins"
	"github.com/deislabs/porter/pkg/printer"
)

// PrintPluginsOptions represent options for the PrintPlugins function
type PrintPluginsOptions struct {
	printer.PrintOptions
}

func (p *Porter) PrintPlugins(opts PrintPluginsOptions) error {
	return errors.New("not implemented")
}

type RunInternalPluginOpts struct {
	Name string
}

func (o *RunInternalPluginOpts) Validate(args []string) error {
	if len(args) == 0 {
		return errors.New("plugin name argument is required")
	}
	if len(args) > 1 {
		return errors.New("multiple plugin name arguments were specified")
	}

	o.Name = args[0]

	return nil
}

func (p *Porter) RunInternalPlugins(opts RunInternalPluginOpts) {
	// TODO: use opts to pick which plugin implementation to us, for now just use the default implementations
	internalPlugins := map[string]plugin.Plugin{
		claimstore.PluginKey: filesystem.NewPlugin(*p.Config),
	}
	plugins.ServeMany(internalPlugins)
}
