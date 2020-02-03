package porter

import (
	"fmt"
	"os"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/pkgmgmt"
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

// PrintPluginOptions represent options for showing a particular plugin.
type PrintPluginOptions struct {
	printer.PrintOptions
	Name string
}

func (o *PrintPluginOptions) Validate(args []string) error {
	err := o.validateName(args)
	if err != nil {
		return err
	}

	return o.ParseFormat()
}

// validateName grabs the name from the first positional argument.
func (o *PrintPluginOptions) validateName(args []string) error {
	switch len(args) {
	case 0:
		return errors.Errorf("no name was specified")
	case 1:
		o.Name = strings.ToLower(args[0])
		return nil
	default:
		return errors.Errorf("only one positional argument may be specified, the name, but multiple were received: %s", args)

	}
}

func (p *Porter) PrintPlugins(opts PrintPluginsOptions) error {
	installedPlugins, err := p.ListPlugins()
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatTable:
		printRow :=
			func(v interface{}) []interface{} {
				m, ok := v.(plugins.Metadata)
				if !ok {
					return nil
				}
				return []interface{}{m.Name, m.VersionInfo.Version, m.VersionInfo.Author}
			}
		return printer.PrintTable(p.Out, installedPlugins, printRow, "Name", "Version", "Author")
	case printer.FormatJson:
		return printer.PrintJson(p.Out, installedPlugins)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, installedPlugins)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) ListPlugins() ([]plugins.Metadata, error) {
	// List out what is installed on the file system
	names, err := p.Plugins.List()
	if err != nil {
		return nil, err
	}

	// Query each plugin and fill out their metadata, handle the
	// cast from the PackageMetadata interface to the concrete type
	installedPlugins := make([]plugins.Metadata, len(names))
	for i, name := range names {
		plugin, err := p.Plugins.GetMetadata(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get version from plugin %s: %s\n ", name, err.Error())
			continue
		}

		meta, _ := plugin.(*plugins.Metadata)
		installedPlugins[i] = *meta
	}

	return installedPlugins, nil
}

func (p *Porter) InstallPlugin(opts plugins.InstallOptions) error {
	err := p.Plugins.Install(opts.InstallOptions)
	if err != nil {
		return err
	}

	plugin, err := p.Plugins.GetMetadata(opts.Name)
	if err != nil {
		return err
	}

	v := plugin.GetVersionInfo()
	fmt.Fprintf(p.Out, "installed %s plugin %s (%s)", opts.Name, v.Version, v.Commit)

	return nil
}

func (p *Porter) UninstallPlugin(opts pkgmgmt.UninstallOptions) error {
	err := p.Plugins.Uninstall(opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "Uninstalled %s plugin", opts.Name)

	return nil
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
		host.PluginKey:       func() plugin.Plugin { return host.NewPlugin() },
	}
}
