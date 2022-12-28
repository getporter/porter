package plugins

import (
	"encoding/json"
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
)

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

		if o.Version != "" {
			return fmt.Errorf("plugin version should not be specified when --file is provided")
		}

		if _, err := cxt.FileSystem.Stat(o.File); err != nil {
			return fmt.Errorf("unable to access --file %s: %w", o.File, err)
		}

		return nil
	}

	return o.InstallOptions.Validate(args)
}

// InstallFileOption is the go representation of plugin installation file format.
type InstallFileOption struct {
	data map[string]pkgmgmt.InstallOptions
	keys []string
}

func (io *InstallFileOption) UnmarshalYAML(unmarshal func(interface{}) error) error {
	data := map[string]pkgmgmt.InstallOptions{}
	err := unmarshal(&data)
	if err != nil {
		return fmt.Errorf("could not unmarshal into plugins.InstallFileOption: %w", err)
	}
	return io.sort(data)
}

func (io *InstallFileOption) UnmarshalJSON(data []byte) error {
	configs := map[string]pkgmgmt.InstallOptions{}
	err := json.Unmarshal(data, &configs)
	if err != nil {
		return fmt.Errorf("could not unmarshal into plugins.InstallFileOption: %w", err)
	}

	return io.sort(configs)
}

// sort returns InstallOptions in alphabetical order.
func (io *InstallFileOption) sort(data map[string]pkgmgmt.InstallOptions) error {

	keys := make([]string, 0, len(data))
	for k, v := range data {
		keys = append(keys, k)
		v.Name = k
		v.PackageType = "plugin"
		data[k] = v
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	io.data = data
	io.keys = keys

	return nil

}

// Configs returns InstallOptions list in alphabetical order.
func (io InstallFileOption) Configs() []pkgmgmt.InstallOptions {
	value := make([]pkgmgmt.InstallOptions, 0, len(io.keys))
	for _, k := range io.keys {
		value = append(value, io.data[k])
	}
	return value
}
