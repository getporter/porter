package porter

import (
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/instance-storage"/filesystem"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/printer"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

// PrintPluginsOptions represent options for the PrintPlugins function
type PrintPluginsOptions struct {
	printer.PrintOptions
}

func (p *Porter) PrintPlugins(opts PrintPluginsOptions) error {
	return errors.New("not implemented")
}

type RunInternalPluginOpts struct {
	Key               string
	selectedPlugin    plugin.Plugin
	selectedInterface string
}

func (o *RunInternalPluginOpts) Validate(args []string, cfg *config.Config) error {
	if len(args) == 0 {
		return errors.New("The positional argument KEY was not specified")
	}
	if len(args) > 1 {
		return errors.New("Multiple positional arguments were specified but only one, KEY is expected")
	}

	o.Key = args[0]

	availableImplementations := getInternalPlugins(cfg)
	selectedPlugin, ok := availableImplementations[o.Key]
	if !ok {
		return errors.Errorf("invalid plugin key specified: %q", o.Key)
	}
	o.selectedPlugin = selectedPlugin()

	parts := strings.Split(o.Key, ".")
	o.selectedInterface = parts[0]

	return nil
}

func (p *Porter) RunInternalPlugins(args []string) {
	// We are not following the normal CLI pattern here because
	// if we write to stdout without the hclog, it will cause the plugin framework to blow up
	var opts RunInternalPluginOpts
	err := opts.Validate(args, p.Config)
	if err != nil {
		logger := hclog.New(&hclog.LoggerOptions{
			Name:   "porter",
			Output: p.Err,
			Level:  hclog.Error,
		})
		logger.Error(err.Error())
		return
	}

	plugins.Serve(opts.selectedInterface, opts.selectedPlugin)
}

func getInternalPlugins(cfg *config.Config) map[string]func() plugin.Plugin {
	return map[string]func() plugin.Plugin{
		filesystem.PluginKey: func() plugin.Plugin { return filesystem.NewPlugin(*cfg) },
	}
}
