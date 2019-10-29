package config

import "errors"

// Data is the data stored in PORTER_HOME/porter.toml|yaml|json
type Data struct {
	// Only define fields here that you need to access from code
	// Values are dynamically applied to flags and don't need to be defined
	InstanceStoragePlugin string          `mapstructure:"instance-storage-plugin"`
	DefaultInstanceStore  string          `mapstructure:"default-instance-store"`
	InstanceStores        []InstanceStore `mapstructure:"instance-store"`
}

type InstanceStore struct {
	Name         string                 `mapstructure:"name"`
	PluginSubkey string                 `mapstructure:"plugin"`
	Config       map[string]interface{} `mapstructure:"config"`
}

func (d *Data) GetInstanceStoragePlugin() string {
	if d == nil || d.InstanceStoragePlugin == "" {
		return "filesystem"
	}

	return d.InstanceStoragePlugin
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
