package porter

import (
	"context"
	"fmt"
	"os"
	"strings"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/printer"
	"github.com/olekukonko/tablewriter"
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
		return fmt.Errorf("no name was specified")
	case 1:
		o.Name = strings.ToLower(args[0])
		return nil
	default:
		return fmt.Errorf("only one positional argument may be specified, the name, but multiple were received: %s", args)

	}
}

func (p *Porter) PrintPlugins(ctx context.Context, opts PrintPluginsOptions) error {
	installedPlugins, err := p.ListPlugins(ctx)
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

func (p *Porter) ListPlugins(ctx context.Context) ([]plugins.Metadata, error) {
	// List out what is installed on the file system
	names, err := p.Plugins.List()
	if err != nil {
		return nil, err
	}

	// Query each plugin and fill out their metadata, handle the
	// cast from the PackageMetadata interface to the concrete type
	installedPlugins := make([]plugins.Metadata, len(names))
	for i, name := range names {
		plugin, err := p.Plugins.GetMetadata(ctx, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get version from plugin %s: %s\n ", name, err.Error())
			continue
		}

		meta, _ := plugin.(*plugins.Metadata)
		installedPlugins[i] = *meta
	}

	return installedPlugins, nil
}

func (p *Porter) ShowPlugin(ctx context.Context, opts ShowPluginOptions) error {
	plugin, err := p.GetPlugin(ctx, opts.Name)
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

func (p *Porter) GetPlugin(ctx context.Context, name string) (*plugins.Metadata, error) {
	meta, err := p.Plugins.GetMetadata(ctx, name)
	if err != nil {
		return nil, err
	}

	plugin, ok := meta.(*plugins.Metadata)
	if !ok {
		return nil, fmt.Errorf("could not cast plugin %s to plugins.Metadata", name)
	}

	return plugin, nil
}

func (p *Porter) InstallPlugin(ctx context.Context, opts plugins.InstallOptions) error {
	err := p.Plugins.Install(ctx, opts.InstallOptions)
	if err != nil {
		return err
	}

	plugin, err := p.Plugins.GetMetadata(ctx, opts.Name)
	if err != nil {
		return fmt.Errorf("failed to get plugin metadata: %w", err)
	}

	v := plugin.GetVersionInfo()
	fmt.Fprintf(p.Out, "installed %s plugin %s (%s)\n", opts.Name, v.Version, v.Commit)

	return nil
}

func (p *Porter) UninstallPlugin(ctx context.Context, opts pkgmgmt.UninstallOptions) error {
	err := p.Plugins.Uninstall(ctx, opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "Uninstalled %s plugin", opts.Name)

	return nil
}
