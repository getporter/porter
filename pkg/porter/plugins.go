package porter

import (
	"context"
	"fmt"
	"os"
	"strings"

	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zapcore"
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
		// First, print the plugin metadata
		fmt.Fprintf(p.Out, "Name: %s\n", plugin.Name)
		fmt.Fprintf(p.Out, "Version: %s\n", plugin.Version)
		fmt.Fprintf(p.Out, "Commit: %s\n", plugin.Commit)
		fmt.Fprintf(p.Out, "Author: %s\n\n", plugin.Author)

		return printer.PrintTable(p.Out, plugin.Implementations, func(v interface{}) []string {
			m, ok := v.(plugins.Implementation)
			if !ok {
				return nil
			}
			return []string{m.Type, m.Name}
		}, "Type", "Implementation")

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
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	installOpts, err := p.getPluginInstallOptions(ctx, opts)
	if err != nil {
		return err
	}
	for _, opt := range installOpts {
		err := p.Plugins.Install(ctx, opt)
		if err != nil {
			return err
		}

		plugin, err := p.Plugins.GetMetadata(ctx, opt.Name)
		if err != nil {
			return fmt.Errorf("failed to get plugin metadata: %w", err)
		}

		v := plugin.GetVersionInfo()
		fmt.Fprintf(p.Out, "installed %s plugin %s (%s)\n", opt.Name, v.Version, v.Commit)
	}

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

func (p *Porter) getPluginInstallOptions(ctx context.Context, opts plugins.InstallOptions) ([]pkgmgmt.InstallOptions, error) {
	_, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	var installConfigs []pkgmgmt.InstallOptions
	if opts.File != "" {
		var data plugins.InstallPluginsSpec
		if log.ShouldLog(zapcore.DebugLevel) {
			// ignoring any error here, printing debug info isn't critical
			contents, _ := p.FileSystem.ReadFile(opts.File)
			log.Debug("read input file", attribute.String("contents", string(contents)))
		}

		if err := encoding.UnmarshalFile(p.FileSystem, opts.File, &data); err != nil {
			return nil, fmt.Errorf("unable to parse %s as an installation document: %w", opts.File, err)
		}

		if err := data.Validate(); err != nil {
			return nil, err
		}

		sortedCfgs := plugins.NewInstallPluginConfigs(data.Plugins)

		for _, config := range sortedCfgs.Values() {
			// if user specified a feed url or mirror using the flags, it will become
			// the default value and apply to empty values parsed from the provided file
			if config.FeedURL == "" {
				config.FeedURL = opts.FeedURL
			}
			if config.Mirror == "" {
				config.Mirror = opts.Mirror
			}

			if err := config.Validate([]string{config.Name}); err != nil {
				return nil, err
			}
			installConfigs = append(installConfigs, config)

		}

		return installConfigs, nil
	}

	installConfigs = append(installConfigs, opts.InstallOptions)
	return installConfigs, nil
}
