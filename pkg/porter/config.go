package porter

import (
	"context"
	"fmt"
	"slices"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/editor"
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
	if slices.Contains(allowedFormats, format) {
		o.Format = format
		return nil
	}

	return fmt.Errorf("invalid format: %s. Allowed formats are: json, yaml, toml", o.RawFormat)
}

// ShowConfig displays the current Porter configuration
func (p *Porter) ShowConfig(ctx context.Context, opts ConfigShowOptions) error {
	_, span := tracing.StartSpan(ctx)
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

// ConfigEditOptions represents options for editing the Porter config
type ConfigEditOptions struct {
	// No special options needed
}

// Validate validates the options for the config edit command
func (o *ConfigEditOptions) Validate(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("config edit does not accept arguments")
	}
	return nil
}

// EditConfig opens the Porter configuration in the user's editor
func (p *Porter) EditConfig(ctx context.Context, opts ConfigEditOptions) error {
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

	// If config doesn't exist, create default
	if !exists {
		err = p.Config.CreateDefaultConfig(ctx, configPath)
		if err != nil {
			return span.Error(fmt.Errorf("error creating default config: %w", err))
		}
	}

	// Read the current config file
	contents, err := p.Config.FileSystem.ReadFile(configPath)
	if err != nil {
		return span.Error(fmt.Errorf("error reading config file: %w", err))
	}

	// Detect the format from the file path
	format := config.DetectConfigFormat(configPath)

	// Determine file extension for temp file
	ext := "." + format
	if format == "yaml" {
		// Use .yaml for consistency even though .yml is also valid
		ext = ".yaml"
	}

	// Open in editor
	ed := editor.New(p.Context, "porter-config"+ext, contents)
	editedContents, err := ed.Run(ctx)
	if err != nil {
		return span.Error(fmt.Errorf("error opening editor: %w", err))
	}

	// Validate the edited content by trying to unmarshal it
	var validatedData config.Data
	err = encoding.Unmarshal(format, editedContents, &validatedData)
	if err != nil {
		return span.Error(fmt.Errorf("invalid config syntax: %w", err))
	}

	// Save the validated content back to the file
	err = p.Config.FileSystem.WriteFile(configPath, editedContents, 0644)
	if err != nil {
		return span.Error(fmt.Errorf("error saving config: %w", err))
	}

	fmt.Fprintf(p.Out, "Configuration saved to %s\n", configPath)

	return nil
}

// ConfigSetOptions represents options for setting a config value
type ConfigSetOptions struct {
	Key   string
	Value string
}

// Validate validates the options for the config set command
func (o *ConfigSetOptions) Validate(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("config set requires exactly 2 arguments: KEY VALUE")
	}
	o.Key = args[0]
	o.Value = args[1]
	return nil
}

// SetConfig sets an individual config value
func (p *Porter) SetConfig(ctx context.Context, opts ConfigSetOptions) error {
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

	var data config.Data

	if !exists {
		// Create default config
		defaultData := config.DefaultDataStore()
		data = defaultData
	} else {
		// Load existing config
		err = encoding.UnmarshalFile(p.Config.FileSystem, configPath, &data)
		if err != nil {
			return span.Error(fmt.Errorf("error loading config from %s: %w", configPath, err))
		}
	}

	// Set the value using the setter
	err = config.SetConfigValue(&data, opts.Key, opts.Value)
	if err != nil {
		return span.Error(err)
	}

	// Update the config data in memory
	p.Config.Data = data

	// Save the config (format is auto-detected from path extension)
	err = p.Config.SaveConfig(ctx, configPath)
	if err != nil {
		return span.Error(fmt.Errorf("error saving config: %w", err))
	}

	fmt.Fprintf(p.Out, "Set %s = %s\n", opts.Key, opts.Value)

	return nil
}
