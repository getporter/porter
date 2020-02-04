package porter

import (
	"fmt"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/secrets/host"
	"get.porter.sh/porter/pkg/storage/filesystem"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

// PrintPluginsOptions represent options for the PrintPlugins function
type PrintPluginsOptions struct {
	printer.PrintOptions
}

func (p *Porter) PrintPlugins(opts PrintPluginsOptions) error {
	installedPlugins, err := p.Plugins.List()
	if err != nil {
		return errors.Wrapf(err, "Failed to get list of installed plugins")
	}

	var pluginsMetadata []plugins.Metadata
	for _, plugin := range installedPlugins {
		metadata, err := p.Plugins.GetMetadata(plugin)
		// lets not break everything just because one plugin failed
		if err != nil {
			if p.Debug {
				fmt.Fprintln(p.Err, "DEBUG Failed to get metadata for ", plugin)
			}
			continue
		}
		pluginsMetadata = append(pluginsMetadata, *metadata)
	}

	implementations := []map[string]string{}

	for _, plugin := range pluginsMetadata {
		if len(plugin.Implementations) != 0 {
			for _, implementation := range plugin.Implementations {
				implementations = append(implementations, map[string]string{
					"Name":           plugin.Name,
					"Type":           implementation.Type,
					"Implementation": implementation.Name,
					"Version":        plugin.Version,
					"Author":         plugin.Author,
				})
			}
		} else {
			// old `plugin version` command don't return implementation details
			implementations = append(implementations, map[string]string{
				"Name":           plugin.Name,
				"Type":           "N/A",
				"Implementation": "N/A",
				"Version":        plugin.Version,
				"Author":         plugin.Author,
			})
		}
	}

	switch opts.Format {
	case printer.FormatTable:
		printMixinRow :=
			func(v interface{}) []interface{} {
				m, ok := v.(map[string]string)
				if !ok {
					return nil
				}
				return []interface{}{m["Name"], m["Type"], m["Implementation"], m["Version"], m["Author"]}
			}
		return printer.PrintTable(p.Out, implementations, printMixinRow, "Name", "Type", "Implementation", "Version", "Author")
	case printer.FormatJson:
		return printer.PrintJson(p.Out, pluginsMetadata)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, pluginsMetadata)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}

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
		host.PluginKey: func() plugin.Plugin { return host.NewPlugin() },
	}
}
