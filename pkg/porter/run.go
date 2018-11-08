package porter

import (
	"fmt"

	"github.com/pkg/errors"
)

func (p *Porter) Run(file string, action Action) error {
	fmt.Fprintln(p.Out, "Starting porter...")

	err := p.Config.LoadManifest(file)
	if err != nil {
		return err
	}

	// validate the configuration for the requested action
	switch action {
	case ActionInstall:
		p.Manifest.Install.Validate(p.Manifest)
	default:
		return errors.Errorf("unsupported action: %q", action)
	}

	// 2. identify and load the mixins

	// 3. identify the bundle action
	// 4. execute the mixin actions

	return nil
}
