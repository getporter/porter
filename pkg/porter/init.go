package porter

import (
	"fmt"

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

	return nil
}
