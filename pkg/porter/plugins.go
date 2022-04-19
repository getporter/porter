package porter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/secrets/plugins/host"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// PrintPluginsOptions represent options for the PrintPlugins function
type PrintPluginsOptions struct {
	printer.PrintOptions
}

// ShowPluginOptions represent options for showing a particular plugin.
type ShowPluginOptions struct {
	printer.PrintOptions
	Name string
}

func (o *ShowPluginOptions) Validate(args []string) error {
	err := o.validateName(args)
	if err != nil {
		return err
	}

	return o.ParseFormat()
}

// validateName grabs the name from the first positional argument.
func (o *ShowPluginOptions) validateName(args []string) error {
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
	case printer.FormatPlaintext:
		printRow :=
			func(v interface{}) []string {
				m, ok := v.(plugins.Metadata)
				if !ok {
					return nil
				}
				return []string{m.Name, m.VersionInfo.Version, m.VersionInfo.Author}
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

func (p *Porter) ShowPlugin(opts ShowPluginOptions) error {
	plugin, err := p.GetPlugin(opts.Name)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatPlaintext:
		// Build and configure our tablewriter
		// TODO: make this a function and reuse it in printer/table.go
		table := tablewriter.NewWriter(p.Out)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: false, Top: true})
		table.SetAutoFormatHeaders(false)

		// First, print the plugin metadata
		fmt.Fprintf(p.Out, "Name: %s\n", plugin.Name)
		fmt.Fprintf(p.Out, "Version: %s\n", plugin.Version)
		fmt.Fprintf(p.Out, "Commit: %s\n", plugin.Commit)
		fmt.Fprintf(p.Out, "Author: %s\n\n", plugin.Author)

		table.SetHeader([]string{"Type", "Implementation"})
		for _, row := range plugin.Implementations {
			table.Append([]string{row.Type, row.Name})
		}
		table.Render()
		return nil

	case printer.FormatJson:
		return printer.PrintJson(p.Out, plugin)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, plugin)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) GetPlugin(name string) (*plugins.Metadata, error) {
	meta, err := p.Plugins.GetMetadata(name)
	if err != nil {
		return nil, err
	}

	plugin, ok := meta.(*plugins.Metadata)
	if !ok {
		return nil, errors.Errorf("could not cast plugin %s to plugins.Metadata", name)
	}

	return plugin, nil
}

func (p *Porter) InstallPlugin(ctx context.Context, opts plugins.InstallOptions) error {
	err := p.Plugins.Install(ctx, opts.InstallOptions)
	if err != nil {
		return err
	}

	plugin, err := p.Plugins.GetMetadata(opts.Name)
	if err != nil {
		return err
	}

	v := plugin.GetVersionInfo()
	fmt.Fprintf(p.Out, "installed %s plugin %s (%s)\n", opts.Name, v.Version, v.Commit)

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
	Key string
}

func (o *RunInternalPluginOpts) Validate(c *portercontext.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("The positional argument KEY was not specified")
	}
	if len(args) > 1 {
		return errors.New("Multiple positional arguments were specified but only one, KEY is expected")
	}

	o.Key = args[0]

	return nil
}

func (o *RunInternalPluginOpts) parsePluginConfig(c *portercontext.Context) (interface{}, error) {
	pluginCfg := map[string]interface{}{}
	if err := json.NewDecoder(c.In).Decode(&pluginCfg); err != nil {
		if err == io.EOF {
			// No plugin config was specified
			return pluginCfg, nil
		}
		return nil, fmt.Errorf("error parsing plugin configuration from stdin as json: %w", err)
	}
	return pluginCfg, nil
}

func (p *Porter) RunInternalPlugins(args []string) error {
	// We are not following the normal CLI pattern here because
	// if we write to stdout without the hclog, it will cause the plugin framework to blow up
	var opts RunInternalPluginOpts
	err := opts.Validate(p.Context, args)
	if err != nil {
		logger := hclog.New(&hclog.LoggerOptions{
			Name:       "porter",
			Output:     p.Err,
			Level:      hclog.Debug,
			JSONFormat: true,
		})
		logger.Error(err.Error())
		return err
	}

	pluginCfg, err := opts.parsePluginConfig(p.Context)
	if err != nil {
		return err
	}

	if _, ok := pluginCfg.(map[string]interface{}); !ok {
		return fmt.Errorf("invalid plugin configuration, should be a document")
	}

	availableImplementations := getInternalPlugins(p.Context, pluginCfg)
	plugins.ServeMany(availableImplementations)
	return nil
}

func getInternalPlugins(c *portercontext.Context, pluginCfg interface{}) map[int]plugin.PluginSet {
	return map[int]plugin.PluginSet{
		// v1 Secrets plugins
		1: {
			host.PluginKey: host.NewPlugin(),
		},
		// v2 Storage plugins
		2: {
			//mongodb.PluginKey:        mongodb.NewPlugin(c, pluginCfg),
			mongodb_docker.PluginKey: mongodb_docker.NewPlugin(c, pluginCfg),
		},
	}
}
