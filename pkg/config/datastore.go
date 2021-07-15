package config

const (
	BuildDriverDocker   = "docker"
	BuildDriverBuildkit = "buildkit"
)

// Data is the data stored in PORTER_HOME/porter.toml|yaml|json.
// Use the accessor functions to ensure default values are handled properly.
type Data struct {
	// Only define fields here that you need to access from code
	// Values are dynamically applied to flags and don't need to be defined

	// BuildDriver is the driver to use when building bundles.
	// Available values are: docker, buildkit.
	BuildDriver string `mapstructure:"build-driver"`

	// DefaultStoragePlugin is the storage plugin to use when no named storage is specified.
	DefaultStoragePlugin string `mapstructure:"default-storage-plugin"`

	// DefaultStorage to use when a named storage is not specified by a flag.
	DefaultStorage string `mapstructure:"default-storage"`

	// ExperimentalFlags is a list of enabled experimental.FeatureFlags.
	// Use Config.IsFeatureEnabled instead of parsing directly.
	ExperimentalFlags []string `mapstructure:"experimental"`

	// CrudStores defined in the configuration file.
	CrudStores []CrudStore `mapstructure:"storage"`

	// DefaultSecretsPlugin is the plugin to use when no plugin is specified.
	DefaultSecretsPlugin string `mapstructure:"default-secrets-plugin"`

	// DefaultSecrets to use when one is not specified by a flag.
	DefaultSecrets string `mapstructure:"default-secrets"`

	// Namespace is the default namespace for commands that do not override it with a flag.
	Namespace string `mapstructure:"namespace"`

	// SecretSources defined in the configuration file.
	SecretSources []SecretSource `mapstructure:"secrets"`
}

// DefaultDataStore used when no config file is found.
func DefaultDataStore() Data {
	return Data{
		BuildDriver:          BuildDriverDocker,
		DefaultStoragePlugin: "filesystem",
		DefaultSecretsPlugin: "host",
	}
}

// SecretSource is the plugin stanza for secrets.
type SecretSource struct {
	PluginConfig `mapstructure:",squash"`
}

// CrudStore is the plugin stanza for storage.
type CrudStore struct {
	PluginConfig `mapstructure:",squash"`
}

// PluginConfig is a standardized config stanza that defines which plugin to
// use and its custom configuration.
type PluginConfig struct {
	Name         string                 `mapstructure:"name"`
	PluginSubKey string                 `mapstructure:"plugin"`
	Config       map[string]interface{} `mapstructure:"config"`
}

func (p PluginConfig) GetName() string {
	return p.Name
}

func (p PluginConfig) GetPluginSubKey() string {
	return p.PluginSubKey
}

func (p PluginConfig) GetConfig() interface{} {
	return p.Config
}
