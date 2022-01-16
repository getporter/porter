package porter

import (
	"fmt"
	"regexp"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/pkg/errors"
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
	err := p.FileSystem.MkdirAll(build.LOCAL_APP, 0700)
	if err != nil {
		return errors.Wrapf(err, "unable to create directory %s", build.LOCAL_APP)
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

	numberRegex := regexp.MustCompile(`\d`)

	if opts.parsedCustoms != nil {
		for k, v := range opts.parsedCustoms {
			if v != "true" && v != "false" && !numberRegex.MatchString(v) {
				v = fmt.Sprintf("\"%s\"", v)
			}
			if err = e.SetValue("custom."+k, v); err != nil {
				return err
			}
		}
	}

	return e.WriteFile(build.LOCAL_MANIFEST)
}
