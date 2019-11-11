package config

import "github.com/pkg/errors"

// Data is the data stored in PORTER_HOME/porter.toml|yaml|json.
// Use the accessor functions to ensure default values are handled properly.
type Data struct {
	// Only define fields here that you need to access from code
	// Values are dynamically applied to flags and don't need to be defined

	// StoragePlugin is the storage plugin to use when no instance store is specified.
	StoragePlugin string `mapstructure:"storage-plugin"`

	// DefaultStorage to use when one is not specified by a flag.
	DefaultStorage string `mapstructure:"default-storage"`

	// CrudStores defined in the configuration file.
	CrudStores []CrudStore `mapstructure:"storage"`

	// SecretsPlugin is the plugin to use when no plugin is specified.
	SecretsPlugin string `mapstructure:"secrets-plugin"`

	// DefaultCredentialSource to use when one is not specified by a flag.
	DefaultSecrets string `mapstructure:"default-secrets"`

	// CredentialSources defined in the configuration file.
	SecretSources []SecretSource `mapstructure:"secrets"`
}

// SecretSource is the plugin stanza for secrets.
type SecretSource struct {
	PluginConfig    `mapstructure:",squash"`
}

// CrudStore is the plugin stanza for storage.
type CrudStore struct {
	PluginConfig     `mapstructure:",squash"`
}

func (d *Data) GetStoragePlugin() string {
	if d == nil || d.StoragePlugin == "" {
		return "filesystem"
	}

	return d.StoragePlugin
}

func (d *Data) GetDefaultStorage() string {
	if d == nil {
		return ""
	}

	return d.DefaultStorage
}

func (d *Data) GetStorage(name string) (CrudStore, error) {
	if d != nil {
		for _, is := range d.CrudStores {
			if is.Name == name {
				return is, nil
			}
		}
	}

	return CrudStore{}, errors.New("store %q not defined")
}

func (d *Data) GetSecretsPlugin() string {
	if d == nil || d.SecretsPlugin == "" {
		return "host"
	}

	return d.SecretsPlugin
}

func (d *Data) GetDefaultSecretSource() string {
	if d == nil {
		return ""
	}

	return d.DefaultSecrets
}

func (d *Data) GetSecretSource(name string) (SecretSource, error) {
	if d != nil {
		for _, cs := range d.SecretSources {
			if cs.Name == name {
				return cs, nil
			}
		}
	}

	return SecretSource{}, errors.New("secrets %q not defined")
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
