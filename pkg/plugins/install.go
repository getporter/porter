package plugins

import (
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
type InstallFileOption map[string]pkgmgmt.InstallOptions

// InstallPluginsConfig is a sorted list of InstallationFileOption in alphabetical order.
type InstallPluginsConfig struct {
	data InstallFileOption
	keys []string
}

// NewInstallPluginConfigs returns a new instance of InstallPluginConfigs with plugins sorted in alphabetical order
// using their names.
func NewInstallPluginConfigs(opt InstallFileOption) InstallPluginsConfig {
	keys := make([]string, 0, len(opt))
	data := make(InstallFileOption, len(opt))
	for k, v := range opt {
		keys = append(keys, k)

		v.Name = k
		v.PackageType = "plugin"
		data[k] = v
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	return InstallPluginsConfig{
		data: data,
		keys: keys,
	}
}

// Configs returns InstallOptions list in alphabetical order.
func (pc InstallPluginsConfig) Configs() []pkgmgmt.InstallOptions {
	value := make([]pkgmgmt.InstallOptions, 0, len(pc.keys))
	for _, k := range pc.keys {
		value = append(value, pc.data[k])
	}
	return value
}
