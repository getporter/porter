package config

import "errors"

// Data is the data stored in PORTER_HOME/porter.toml|yaml|json
type Data struct {
	// Only define fields here that you need to access from code
	// Values are dynamically applied to flags and don't need to be defined
	StoragePlugin        string          `mapstructure:"storage-plugin"`
	DefaultInstanceStore string          `mapstructure:"default-instance-store"`
	InstanceStores       []InstanceStore `mapstructure:"instance-store"`
}

type InstanceStore struct {
	Name         string                 `mapstructure:"name"`
	PluginSubKey string                 `mapstructure:"plugin"`
	Config       map[string]interface{} `mapstructure:"config"`
}

func (is InstanceStore) GetName() string {
	return is.Name
}

func (is InstanceStore) GetPluginSubKey() string {
	return is.PluginSubKey
}

func (is InstanceStore) GetConfig() interface{} {
	return is.Config
}

func (d *Data) GetStoragePlugin() string {
	if d == nil || d.StoragePlugin == "" {
		return "filesystem"
	}

	return d.StoragePlugin
}

func (d *Data) GetDefaultInstanceStore() string {
	if d == nil {
		return ""
	}

	return d.DefaultInstanceStore
}

func (d *Data) GetInstanceStore(name string) (InstanceStore, error) {
	if d != nil {
		for _, is := range d.InstanceStores {
			if is.Name == name {
				return is, nil
			}
		}
	}

	return InstanceStore{}, errors.New("instance-store %q not defined")
}

var _ DataStoreLoaderFunc = NoopDataLoader

// NoopDataLoader skips loading the datastore.
func NoopDataLoader(config *Config) error {
	return nil
}
