package porter

import (
	"context"
	"fmt"
	"path/filepath"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/editor"
	"get.porter.sh/porter/pkg/tracing"
)

// ConfigShowOptions are the options for the ConfigShow command.
type ConfigShowOptions struct{}

// ConfigEditOptions are the options for the ConfigEdit command.
type ConfigEditOptions struct{}

// defaultConfigTemplate is the default TOML template for a new config file.
const defaultConfigTemplate = `# Porter configuration
# https://porter.sh/configuration/

# verbosity = "info"
# namespace = ""
# default-storage-plugin = "mongodb-docker"
# default-secrets-plugin = "host"
`

// GetConfigFilePath returns the path to the porter config file.
// It checks for config files with extensions: toml, yaml, yml, json, hcl.
// Returns the path, whether the file exists, and any error.
func (p *Porter) GetConfigFilePath() (string, bool, error) {
	home, err := p.GetHomeDir()
	if err != nil {
		return "", false, fmt.Errorf("could not get porter home directory: %w", err)
	}

	extensions := []string{"toml", "yaml", "yml", "json", "hcl"}
	for _, ext := range extensions {
		path := filepath.Join(home, "config."+ext)
		exists, err := p.FileSystem.Exists(path)
		if err != nil {
			return "", false, fmt.Errorf("could not check if config file exists: %w", err)
		}
		if exists {
			return path, true, nil
		}
	}

	// Default to toml if no config file exists
	return filepath.Join(home, "config.toml"), false, nil
}

// ConfigShow displays the contents of the porter config file.
func (p *Porter) ConfigShow(ctx context.Context, opts ConfigShowOptions) error {
	_, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	path, exists, err := p.GetConfigFilePath()
	if err != nil {
		return span.Error(err)
	}

	if !exists {
		fmt.Fprintln(p.Out, "No configuration file found.")
		fmt.Fprintln(p.Out, "Use 'porter config edit' to create one.")
		return nil
	}

	contents, err := p.FileSystem.ReadFile(path)
	if err != nil {
		return span.Error(fmt.Errorf("could not read config file %s: %w", path, err))
	}

	fmt.Fprintln(p.Out, string(contents))
	return nil
}

// ConfigEdit opens the porter config file in an editor.
// If the file does not exist, it creates a default template first.
func (p *Porter) ConfigEdit(ctx context.Context, opts ConfigEditOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	path, exists, err := p.GetConfigFilePath()
	if err != nil {
		return span.Error(err)
	}

	var contents []byte
	if exists {
		contents, err = p.FileSystem.ReadFile(path)
		if err != nil {
			return span.Error(fmt.Errorf("could not read config file %s: %w", path, err))
		}
	} else {
		contents = []byte(defaultConfigTemplate)
	}

	ed := editor.New(p.Context, "porter-config.toml", contents)
	output, err := ed.Run(ctx)
	if err != nil {
		return span.Error(fmt.Errorf("unable to open editor: %w", err))
	}

	// Ensure the directory exists
	dir := filepath.Dir(path)
	if err := p.FileSystem.MkdirAll(dir, pkg.FileModeDirectory); err != nil {
		return span.Error(fmt.Errorf("could not create config directory %s: %w", dir, err))
	}

	if err := p.FileSystem.WriteFile(path, output, pkg.FileModeWritable); err != nil {
		return span.Error(fmt.Errorf("could not write config file %s: %w", path, err))
	}

	return nil
}
