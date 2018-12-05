package porter

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin"
)

func (p *Porter) Run(file string, action config.Action) error {
	fmt.Fprintf(p.Out, "executing porter %s configuration from %s\n", action, file)

	err := p.Config.LoadManifestFrom(file)
	if err != nil {
		return err
	}

	steps, err := p.Manifest.GetSteps(action)
	if err != nil {
		return err
	}

	err = steps.Validate(p.Manifest)
	if err != nil {
		return errors.Wrap(err, "invalid action configuration")
	}

	mixinsDir, err := p.GetMixinsDir()
	if err != nil {
		return err
	}

	for _, step := range steps {
		err := p.Config.Manifest.ResolveStep(step)
		if err != nil {
			return errors.Wrap(err, "unable to resolve sourced values")
		}
		runner := p.loadRunner(step, action, mixinsDir)

		err = runner.Validate()
		if err != nil {
			return errors.Wrap(err, "mixin validation failed")
		}

		fmt.Fprintln(p.Out, step.Description)
		err = runner.Run()
		if err != nil {
			return errors.Wrap(err, "mixin execution failed")
		}
	}

	fmt.Fprintln(p.Out, "execution completed successfully!")
	return nil
}

func (p *Porter) loadRunner(s *config.Step, action config.Action, mixinsDir string) *mixin.Runner {
	name := s.GetMixinName()
	mixinDir := filepath.Join(mixinsDir, name)

	r := mixin.NewRunner(name, mixinDir, true)
	r.Command = string(action)
	r.Context = p.Context

	stepBytes, _ := yaml.Marshal(s)
	r.Step = string(stepBytes)

	return r
}
