package kubernetes

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

type UninstallAction struct {
	Steps []UninstallStep `yaml:"uninstall"`
}

type UninstallStep struct {
	UninstallArguments `yaml:"kubernetes"`
}

type UninstallArguments struct {
	Step `yaml:",inline"`

	Namespace string   `yaml:"namespace"`
	Manifests []string `yaml:"manifests,omitempty"`

	Force       *bool  `yaml:force,omitempty"`
	GracePeriod *int   `yaml:"gracePeriod,omitempty"`
	Selector    string `yaml:"selector,omitempty"`
	Timeout     *int   `yaml:"timeout,omitempty"`
	Wait        *bool  `yaml:"wait,omitempty"`
}

// Uninstall will delete anything created during the install or upgrade step
func (m *Mixin) Uninstall() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var action UninstallAction
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
		commandPayload, err := m.buildUninstallCommand(step.UninstallArguments, manifestPath)
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
	return nil
}

func (m *Mixin) buildUninstallCommand(args UninstallArguments, manifestPath string) ([]string, error) {
	command := []string{"delete", "-f", manifestPath}
	if args.Namespace != "" {
		command = append(command, "-n", args.Namespace)
	}

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
				// default to zero
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

	if args.Selector != "" {
		command = append(command, fmt.Sprintf("--selector=%s", args.Selector))
	}

	if args.Timeout != nil {
		timeout := *args.Timeout
		command = append(command, fmt.Sprintf("--timeout=%ds", timeout))
	}

	waitForIt := true
	if args.Wait != nil {
		waitForIt = *args.Wait
	}
	if waitForIt {
		command = append(command, "--wait")
	}

	return command, nil
}
