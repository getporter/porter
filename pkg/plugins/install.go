package plugins

import (
	"fmt"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
)

// SchemaTypePlugins is the default schemaType value for InstallPluginsSpec resources
const SchemaTypePlugins = "Plugins"

// InstallPluginsSchemaVersion represents the version associated with the schema
// plugins configuration documents.
var InstallPluginsSchemaVersion = cnab.SchemaVersion("1.0.0")

type InstallOptions struct {
	pkgmgmt.InstallOptions

	File string
}

func (o *InstallOptions) Validate(args []string, cxt *portercontext.Context) error {
	o.PackageType = "plugin"
	if o.File != "" {
		if len(args) > 0 {
			return fmt.Errorf("plugin name should not be specified when --file is provided")
		}

		if o.URL != "" {
			return fmt.Errorf("plugin URL should not be specified when --file is provided")
		}

		// version should not be set to anything other than the default value
		if o.Version != "" && o.Version != "latest" {
			return fmt.Errorf("plugin version %s should not be specified when --file is provided", o.Version)
		}

		if _, err := cxt.FileSystem.Stat(o.File); err != nil {
			return fmt.Errorf("unable to access --file %s: %w", o.File, err)
		}

		return nil
	}

	return o.InstallOptions.Validate(args)
}

// InstallPluginsSpec represents the user-defined configuration for plugins installation.
type InstallPluginsSpec struct {
	SchemaType    string               `yaml:"schemaType"`
	SchemaVersion string               `yaml:"schemaVersion"`
	Plugins       InstallPluginsConfig `yaml:"plugins"`
}

func (spec InstallPluginsSpec) Validate() error {
	if spec.SchemaType == "" {
		// Default the schema type before importing into the database if it's not set already
		// SchemaType isn't really used by our code, it's a type hint for editors, but this will ensure we are consistent in our persisted documents
		spec.SchemaType = SchemaTypePlugins
	} else if !strings.EqualFold(spec.SchemaType, SchemaTypePlugins) {
		return fmt.Errorf("invalid schemaType %s, expected %s", spec.SchemaType, SchemaTypePlugins)
	}

	if InstallPluginsSchemaVersion != schema.Version(spec.SchemaVersion) {
		if spec.SchemaVersion == "" {
			spec.SchemaVersion = "(none)"
		}
		return fmt.Errorf("invalid schemaVersion provided: %s. This version of Porter is compatible with %s.", spec.SchemaVersion, InstallPluginsSchemaVersion)
	}
	return nil
}

// InstallPluginsConfig is the go representation of plugin installation file format.
type InstallPluginsConfig map[string]pkgmgmt.InstallOptions

// InstallPluginsConfigList is a sorted list of InstallationFileOption in alphabetical order.
type InstallPluginsConfigList struct {
	data InstallPluginsConfig
	keys []string
}

// NewInstallPluginConfigs returns a new instance of InstallPluginConfigs with plugins sorted in alphabetical order
// using their names.
func NewInstallPluginConfigs(opt InstallPluginsConfig) InstallPluginsConfigList {
	keys := make([]string, 0, len(opt))
	data := make(InstallPluginsConfig, len(opt))
	for k, v := range opt {
		keys = append(keys, k)

		v.Name = k
		v.PackageType = "plugin"
		data[k] = v
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return InstallPluginsConfigList{
		data: data,
		keys: keys,
	}
}

// Values returns InstallOptions list in alphabetical order.
func (pc InstallPluginsConfigList) Values() []pkgmgmt.InstallOptions {
	value := make([]pkgmgmt.InstallOptions, 0, len(pc.keys))
	for _, k := range pc.keys {
		value = append(value, pc.data[k])
	}
	return value
}
