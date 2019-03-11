package porter

import (
	"fmt"
	"path/filepath"

	"github.com/deislabs/porter/pkg/config"
	"github.com/pkg/errors"
)

func (p *Porter) Create() error {
	fmt.Fprintln(p.Out, "creating porter configuration in the current directory")

	configTmpl, err := p.GetManifestTemplate()
	if err != nil {
		return err
	}

	err = p.FileSystem.WriteFile(config.Name, configTmpl, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s", config.Name)
	}

	runTmpl, err := p.GetRunScriptTemplate()
	if err != nil {
		return err
	}

	err = p.FileSystem.MkdirAll(filepath.Dir(config.RunScript), 0755)
	if err != nil {
		return err
	}

	err = p.FileSystem.WriteFile(config.RunScript, runTmpl, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s", config.RunScript)
	}

	return nil
}
