package porter

import (
	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/pkg/errors"
)

// metadataOpts contain manifest fields eligible for dynamic
// updating prior to saving Porter's internal version of the manifest
type metadataOpts struct {
	Name    string
	Version string
	Tag     string // This may be set via Publish
}

// generateInternalManifest decodes the manifest designated by filepath and applies
// the provided generateInternalManifestOpts, saving the updated manifest to the path
// designated by build.LOCAL_MANIFEST
func (p *Porter) generateInternalManifest(opts metadataOpts) error {
	// Create the local app dir if it does not already exist
	err := p.FileSystem.MkdirAll(build.LOCAL_APP, 0755)
	if err != nil {
		return errors.Wrapf(err, "unable to create directory %s", build.LOCAL_APP)
	}

	e := manifest.NewEditor(p.Context)
	err = e.ReadFile(config.Name)
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

	return e.WriteFile(build.LOCAL_MANIFEST)
}
