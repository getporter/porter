package kubernetes

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type ExecuteAction struct {
	Steps []ExecuteStep // using UnmarshalYAML so that we don't need a custom type per action
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - kubernetes: ...
// and puts the steps into the Action.Steps field
func (a *ExecuteAction) UnmarshalYAML(unmarshal func(interface{}) error) error {
	actionMap := map[interface{}][]interface{}{}
	err := unmarshal(&actionMap)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal yaml into an action map of kubernetes steps")
	}

	for _, stepMaps := range actionMap {
		b, err := yaml.Marshal(stepMaps)
		if err != nil {
			return err
		}

		var steps []ExecuteStep
		err = yaml.Unmarshal(b, &steps)
		if err != nil {
			return err
		}

		a.Steps = append(a.Steps, steps...)
	}

	return nil
}

type ExecuteStep struct {
	ExecuteInstruction `yaml:"kubernetes"`
}

type ExecuteInstruction struct {
	InstallArguments `yaml:",inline"`

	// Upgrade specific arguments
	Force       *bool `yaml:"force,omitempty"`
	GracePeriod *int  `yaml:"gracePeriod,omitempty"`
	Overwrite   *bool `yaml:"overwrite,omitempty"`
	Prune       *bool `yaml:"prune,omitempty"`
	Timeout     *int  `yaml:"timeout,omitempty"`
}

// Execute will reapply manifests using kubectl
func (m *Mixin) Execute() error {

	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var action ExecuteAction
	err = yaml.Unmarshal(payload, &action)
	if err != nil {
		return err
	}

	if len(action.Steps) != 1 {
		return errors.Errorf("expected a single step, but got %d", len(action.Steps))
	}

	step := action.Steps[0]
	var commands []*exec.Cmd

	for _, manifestPath := range step.Manifests {
		commandPayload, err := m.buildExecuteCommand(step.ExecuteInstruction, manifestPath)
		if err != nil {
			return err
		}
		cmd := m.NewCommand("kubectl", commandPayload...)
		commands = append(commands, cmd)
	}

	for _, cmd := range commands {
		cmd.Stdout = m.Out
		cmd.Stderr = m.Err

		err = cmd.Start()
		if err != nil {
			prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
			return errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
		}
		err = cmd.Wait()
		if err != nil {
			prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
			return errors.Wrap(err, fmt.Sprintf("error running command %s", prettyCmd))
		}
	}

	err = m.handleOutputs(step.Outputs)
	return err
}

func (m *Mixin) buildExecuteCommand(args ExecuteInstruction, manifestPath string) ([]string, error) {
	command, err := m.buildInstallCommand(args.InstallArguments, manifestPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create upgrade command")
	}

	// // Upgrade specific arguments
	// Timeout     *int  `yaml:"timeout,omitempty"`

	if args.Force != nil {
		forceIt := *args.Force
		if forceIt {
			command = append(command, "--force")
			if args.GracePeriod != nil {
				gracePeriod := *args.GracePeriod
				if gracePeriod != 0 {
					return nil, fmt.Errorf("grace period must be zero when force is specified: %d", gracePeriod)
				}
			} else {
				//set the grace period to zero
				command = append(command, "--grace-period=0")
			}

		}
	}

	if args.GracePeriod != nil {
		gracePeriod := *args.GracePeriod
		command = append(command, fmt.Sprintf("--grace-period=%d", gracePeriod))
		if gracePeriod == 0 {
			command = append(command, "--force")
		}
	}

	if args.Prune != nil {
		pruneIt := *args.Prune
		if pruneIt {
			command = append(command, "--prune=true")
		}
	}

	if args.Overwrite != nil {
		overwriteIt := *args.Overwrite
		if !overwriteIt {
			command = append(command, "--overwrite=false")
		}
	}

	if args.Timeout != nil {
		timeout := *args.Timeout
		command = append(command, fmt.Sprintf("--timeout=%ds", timeout))
	}

	return command, nil
}
