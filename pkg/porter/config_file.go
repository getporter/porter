package porter

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/editor"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/spf13/viper"
)

// ConfigShowOptions are the options for the ConfigShow command.
type ConfigShowOptions struct{}

// ConfigEditOptions are the options for the ConfigEdit command.
type ConfigEditOptions struct{}

// defaultConfigTemplate is the default YAML template for a new config file.
const defaultConfigTemplate = `schemaVersion: "2.0.0"
current-context: default
contexts:
  - name: default
    config:
      # verbosity: "info"
      # namespace: ""
      # default-storage-plugin: "mongodb-docker"
      # default-secrets-plugin: "host"
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

	// Default to yaml if no config file exists
	return filepath.Join(home, "config.yaml"), false, nil
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

	ed := editor.New(p.Context, "porter-config.yaml", contents)
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

// ConfigContextList lists all contexts in the porter config file.
func (p *Porter) ConfigContextList(ctx context.Context) error {
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

	v := viper.New()
	v.SetFs(p.FileSystem)
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return span.Error(fmt.Errorf("could not read config file %s: %w", path, err))
	}

	rawMap := v.AllSettings()
	if _, isMultiContext := rawMap["schemaversion"]; !isMultiContext {
		fmt.Fprintln(p.Out, "Config file is in legacy flat format with a single implicit context.")
		return nil
	}

	// Determine the active context: flag/env > current-context in file > "default"
	active := p.Config.ContextName
	if active == "" {
		if cc, _ := rawMap["current-context"].(string); cc != "" {
			active = cc
		}
	}
	if active == "" {
		active = "default"
	}

	contexts, _ := rawMap["contexts"].([]interface{})
	for _, c := range contexts {
		ctxMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := ctxMap["name"].(string)
		if name == "" {
			continue
		}
		marker := "  "
		if name == active {
			marker = "* "
		}
		fmt.Fprintf(p.Out, "%s%s\n", marker, name)
	}
	return nil
}

// migrationHeader is prepended to the existing config content when migrating
// from the legacy flat format to the multi-context format.
const migrationHeader = `schemaVersion: "2.0.0"
current-context: default
contexts:
  - name: default
    config:
`

// ConfigMigrate converts a legacy flat config file to the multi-context format.
// The existing settings are wrapped in a context named "default".
// Only YAML config files are supported; other formats return an error with
// guidance for manual migration.
func (p *Porter) ConfigMigrate(ctx context.Context) error {
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

	if bytes.Contains(contents, []byte("schemaVersion:")) {
		fmt.Fprintln(p.Out, "Config file is already using the multi-context format.")
		return nil
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	if ext != "yaml" && ext != "yml" {
		return span.Error(fmt.Errorf(
			"automatic migration is only supported for YAML config files; "+
				"your %s config must be migrated manually by wrapping your settings under:\n\n"+
				"  schemaVersion: \"2.0.0\"\n"+
				"  current-context: default\n"+
				"  contexts:\n"+
				"    - name: default\n"+
				"      config:\n"+
				"        # your existing settings here",
			ext,
		))
	}

	// Normalize line endings, then indent each non-empty line by 6 spaces so
	// the existing settings sit correctly under "contexts[0].config".
	contents = bytes.ReplaceAll(contents, []byte("\r\n"), []byte("\n"))
	indented := indentLines(bytes.TrimRight(contents, "\n"), 6)
	migrated := append([]byte(migrationHeader), indented...)
	migrated = append(migrated, '\n')

	if err := p.FileSystem.WriteFile(path, migrated, pkg.FileModeWritable); err != nil {
		return span.Error(fmt.Errorf("could not write migrated config file %s: %w", path, err))
	}

	fmt.Fprintf(p.Out, "Migrated %s to multi-context format. Use 'porter config show' to review.\n", filepath.Base(path))
	return nil
}

// indentLines prepends spaces to every non-empty line in content.
func indentLines(content []byte, spaces int) []byte {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = prefix + line
		}
	}
	return []byte(strings.Join(lines, "\n"))
}

// currentContextRe matches the current-context line in a YAML config file.
var currentContextRe = regexp.MustCompile(`(?m)^current-context:.*$`)

// ConfigContextUse sets the current-context in the porter config file.
func (p *Porter) ConfigContextUse(ctx context.Context, name string) error {
	_, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	path, exists, err := p.GetConfigFilePath()
	if err != nil {
		return span.Error(err)
	}
	if !exists {
		return span.Error(fmt.Errorf("no config file found; use 'porter config edit' to create one"))
	}

	contents, err := p.FileSystem.ReadFile(path)
	if err != nil {
		return span.Error(fmt.Errorf("could not read config file %s: %w", path, err))
	}

	if !bytes.Contains(contents, []byte("schemaVersion: \""+config.ConfigSchemaVersion+"\"")) {
		return span.Error(fmt.Errorf("config file is not a versioned multi-context file (schemaVersion: %q required)", config.ConfigSchemaVersion))
	}

	replacement := []byte("current-context: " + name)
	if currentContextRe.Match(contents) {
		contents = currentContextRe.ReplaceAll(contents, replacement)
	} else {
		// Insert after the schemaVersion line
		contents = bytes.Replace(contents,
			[]byte("schemaVersion: \""+config.ConfigSchemaVersion+"\""),
			[]byte("schemaVersion: \""+config.ConfigSchemaVersion+"\"\ncurrent-context: "+name),
			1)
	}

	if err := p.FileSystem.WriteFile(path, contents, pkg.FileModeWritable); err != nil {
		return span.Error(fmt.Errorf("could not write config file %s: %w", path, err))
	}

	fmt.Fprintf(p.Out, "Switched to context %q.\n", name)
	return nil
}
