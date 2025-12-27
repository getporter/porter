package porter

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/tracing"
)

var (
	ConfigShowAllowedFormats = []printer.Format{printer.FormatJson, printer.FormatYaml, "toml"}
	ConfigShowDefaultFormat  = "toml"
)

// ConfigShowOptions represents options for showing the Porter config
type ConfigShowOptions struct {
	printer.PrintOptions
}

// Validate validates the options for the config show command
func (o *ConfigShowOptions) Validate(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("config show does not accept arguments")
	}

	// Convert format string to printer.Format for validation
	allowedFormats := ConfigShowAllowedFormats
	defaultFormat := ConfigShowDefaultFormat

	// Default unspecified format
	if o.RawFormat == "" {
		o.RawFormat = string(defaultFormat)
	}

	// Validate format
	format := printer.Format(o.RawFormat)
	for _, f := range allowedFormats {
		if f == format {
			o.Format = format
			return nil
		}
	}

	return fmt.Errorf("invalid format: %s. Allowed formats are: json, yaml, toml", o.RawFormat)
}

// ShowConfig displays the current Porter configuration
func (p *Porter) ShowConfig(ctx context.Context, opts ConfigShowOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// Get the config path
	configPath, err := p.Config.GetConfigPath()
	if err != nil {
		return span.Error(err)
	}

	// Check if config exists
	exists, err := p.Config.FileSystem.Exists(configPath)
	if err != nil {
		return span.Error(fmt.Errorf("error checking if config exists: %w", err))
	}

	var data *config.Data
	var outputFormat string

	if !exists {
		// Use default config
		defaultData := config.DefaultDataStore()
		data = &defaultData
		// Use format from options
		outputFormat = string(opts.Format)
	} else {
		// Load config from file to ensure we have the latest values
		var loadedData config.Data
		err = encoding.UnmarshalFile(p.Config.FileSystem, configPath, &loadedData)
		if err != nil {
			return span.Error(fmt.Errorf("error loading config from %s: %w", configPath, err))
		}
		data = &loadedData

		// Determine output format: use opts.Format if explicitly set,
		// otherwise use the existing file format
		if opts.RawFormat != "" && opts.RawFormat != string(ConfigShowDefaultFormat) {
			outputFormat = string(opts.Format)
		} else {
			outputFormat = config.DetectConfigFormat(configPath)
		}
	}

	// Marshal to requested format
	output, err := encoding.Marshal(outputFormat, data)
	if err != nil {
		return span.Error(fmt.Errorf("error marshaling config: %w", err))
	}

	// Write to stdout
	fmt.Fprintln(p.Out, string(output))

	return nil
}
