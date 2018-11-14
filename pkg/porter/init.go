package porter

import (
	"fmt"
	"path/filepath"

	"github.com/deislabs/porter/pkg/config"
	"github.com/pkg/errors"
)

func (p *Porter) Init() error {
	fmt.Fprintln(p.Out, "initializing porter configuration in the current directory")

	configTmpl, err := p.Config.GetPorterConfigTemplate()
	if err != nil {
		return err
	}

	err = p.FileSystem.WriteFile(config.Name, configTmpl, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s", config.Name)
	}

	runTmpl, err := p.Config.GetRunScriptTemplate()
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
