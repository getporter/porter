package porter

import (
	"fmt"
	"io/ioutil"

	"github.com/deislabs/porter/pkg/config"
	"github.com/pkg/errors"
)

func (p *Porter) Init() error {
	fmt.Fprintln(p.Out, "initializing porter configuration in the current directory")

	configTmpl, err := config.GetPorterConfigTemplate()
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(config.Name, configTmpl, 0644)
	if err != nil {
		return errors.Wrapf(err, "failed to write %s", config.Name)
	}

	return nil
}
