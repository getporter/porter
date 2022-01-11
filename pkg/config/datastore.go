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
	// Do not use directly, use Config.GetBuildDriver.
	BuildDriver string `mapstructure:"build-driver"`

	// RuntimeDriver is the driver to use when executing bundles.
	// Available values are: docker, kubernetes.
	RuntimeDriver string `mapstructure:"runtime-driver"`

	// AllowDockerHostAccess grants bundles access to the underlying docker host
	// upon which it is running so that it can do things like build and run containers.
	// It's a security risk.
	AllowDockerHostAccess bool `mapstructure:"allow-docker-host-access"`

	// DefaultStoragePlugin is the storage plugin to use when no named storage is specified.
	DefaultStoragePlugin string `mapstructure:"default-storage-plugin"`

	// DefaultStorage to use when a named storage is not specified by a flag.
	DefaultStorage string `mapstructure:"default-storage"`

	// ExperimentalFlags is a list of enabled experimental.FeatureFlags.
	// Use Config.IsFeatureEnabled instead of parsing directly.
	ExperimentalFlags []string `mapstructure:"experimental"`

	// StoragePlugins defined in the configuration file.
	StoragePlugins []StoragePlugin `mapstructure:"storage"`

	// DefaultSecretsPlugin is the plugin to use when no plugin is specified.
	DefaultSecretsPlugin string `mapstructure:"default-secrets-plugin"`

	// DefaultSecrets to use when one is not specified by a flag.
	DefaultSecrets string `mapstructure:"default-secrets"`

	// Namespace is the default namespace for commands that do not override it with a flag.
	Namespace string `mapstructure:"namespace"`

	// SecretsPlugin defined in the configuration file.
	SecretsPlugin []SecretsPlugin `mapstructure:"secrets"`

	// Logs are settings related to Porter's log files.
	Logs LogConfig `mapstructure:"logs"`

	// Telemetry are settings related to Porter's tracing with open telemetry.
	Telemetry TelemetryConfig `mapstructure:telemetry`
}

// DefaultDataStore used when no config file is found.
func DefaultDataStore() Data {
	return Data{
		BuildDriver:          BuildDriverDocker,
		DefaultStoragePlugin: "mongodb-docker",
		DefaultSecretsPlugin: "host",
		Logs:                 LogConfig{Level: "info"},
	}
}

// SecretsPlugin is the plugin stanza for secrets.
type SecretsPlugin struct {
	PluginConfig `mapstructure:",squash"`
}

// StoragePlugin is the plugin stanza for storage.
type StoragePlugin struct {
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
