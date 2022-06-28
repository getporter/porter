package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/yaml"
)

// metadataOpts contain manifest fields eligible for dynamic
// updating prior to saving Porter's internal version of the manifest
type metadataOpts struct {
	Name    string
	Version string
}

// generateInternalManifest decodes the manifest designated by filepath and applies
// the provided generateInternalManifestOpts, saving the updated manifest to the path
// designated by build.LOCAL_MANIFEST
func (p *Porter) generateInternalManifest(opts BuildOptions) error {
	// Create the local app dir if it does not already exist
	err := p.FileSystem.MkdirAll(build.LOCAL_APP, pkg.FileModeDirectory)
	if err != nil {
		return fmt.Errorf("unable to create directory %s: %w", build.LOCAL_APP, err)
	}

	e := yaml.NewEditor(p.Context)
	err = e.ReadFile(opts.File)
	if err != nil {
		return err
	}

	if opts.Name != "" {
		if err = e.SetValue("name", opts.Name); err != nil {
			return err
		}
	}

	if opts.Version != "" {
		if err = e.SetValue("version", opts.Version); err != nil {
			return err
		}
	}

	for k, v := range opts.parsedCustoms {
		if err = e.SetValue("custom."+k, v); err != nil {
			return err
		}
	}

	return e.WriteFile(build.LOCAL_MANIFEST)
}
