package kubernetes

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type InstallAction struct {
	Steps []InstallStep `yaml:"install"`
}

type InstallStep struct {
	InstallArguments `yaml:"kubernetes"`
}

type InstallArguments struct {
	Step `yaml:",inline"`

	Namespace string   `yaml:"namespace"`
	Manifests []string `yaml:"manifests,omitempty"`
	Record    *bool    `yaml:"record,omitempty"`
	Selector  string   `yaml:"selector,omitempty"`
	Validate  *bool    `yaml:"validate,omitempty"`
	Wait      *bool    `yaml:"wait,omitempty"`
}

func (m *Mixin) Install() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}
	var action InstallAction
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
		commandPayload, err := m.buildInstallCommand(step.InstallArguments, manifestPath)
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
			prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
			return errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
		}
		err = cmd.Wait()
		if err != nil {
			prettyCmd := fmt.Sprintf("%s%s", cmd.Dir, strings.Join(cmd.Args, " "))
			return errors.Wrap(err, fmt.Sprintf("error running command %s", prettyCmd))
		}
	}

	err = m.handleOutputs(step.Outputs)
	return err
}

func (m *Mixin) getInstallStep(payload []byte) (*InstallStep, error) {
	var step InstallStep
	err := yaml.Unmarshal(payload, &step)
	if err != nil {
		return nil, err
	}

	return &step, nil
}

func (m *Mixin) buildInstallCommand(step InstallArguments, manifestPath string) ([]string, error) {
	command := []string{"apply", "-f", manifestPath}
	if step.Namespace != "" {
		command = append(command, "-n", step.Namespace)
	}

	if step.Record != nil {
		recordIt := *step.Record
		if recordIt {
			command = append(command, "--record=true")
		}
	}

	if step.Selector != "" {
		command = append(command, fmt.Sprintf("--selector=%s", step.Selector))
	}

	if step.Validate != nil {
		validateIt := *step.Validate
		if !validateIt {
			command = append(command, "--validate=false")
		}
	}

	waitForIt := true
	if step.Wait != nil {
		waitForIt = *step.Wait
	}
	if waitForIt {
		command = append(command, "--wait")
	}

	return command, nil
}
