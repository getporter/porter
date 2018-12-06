package porter

import (
	"fmt"
	"path/filepath"
	"strings"

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

	err = p.FileSystem.MkdirAll(mixin.OutputsDir, 0755)
	if err != nil {
		return errors.Wrapf(err, "could not create outputs directory %s", mixin.OutputsDir)
	}

	for _, step := range steps {
		err := p.Manifest.ResolveStep(step)
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

		err = p.collectStepOutput(step)
		if err != nil {
			return err
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

func (p *Porter) collectStepOutput(step *config.Step) error {
	outputs, err := p.readOutputs()
	if err != nil {
		return err
	}
	return p.Manifest.ApplyOutputs(step, outputs)
}

func (p *Porter) readOutputs() ([]string, error) {
	var outputs []string

	outfiles, err := p.FileSystem.ReadDir(mixin.OutputsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list %s", mixin.OutputsDir)
	}

	for _, outfile := range outfiles {
		if outfile.IsDir() {
			continue
		}

		outpath := filepath.Join(mixin.OutputsDir, outfile.Name())
		contents, err := p.FileSystem.ReadFile(outpath)
		if err != nil {
			return nil, errors.Wrapf(err, "could not read output file %s", outpath)
		}

		for _, line := range strings.Split(string(contents), "\n") {
			// remove any empty lines from the split process
			if len(line) > 0 {
				outputs = append(outputs, line)
			}
		}
		// remove file when we have read it, it shouldn't be here for the
		// next step
		err = p.FileSystem.Remove(outpath)
		if err != nil {
			return nil, err
		}
	}

	return outputs, nil
}
