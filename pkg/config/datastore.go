package config

// Data is the data stored in PORTER_HOME/porter.toml|yaml|json
type Data struct {
	// Only define fields here that you need to access from code
	// Values are dynamically applied to flags and don't need to be defined
	InstanceStoragePlugin string `mapstructure:"instance-storage-plugin"`
}

func (d *Data) GetInstanceStoragePlugin() string {
	if d == nil || d.InstanceStoragePlugin == "" {
		return "filesystem"
	}

	return d.InstanceStoragePlugin
}

var _ DataStoreLoaderFunc = NoopDataLoader

// NoopDataLoader skips loading the datastore.
func NoopDataLoader(config *Config) error {
	return nil
}
