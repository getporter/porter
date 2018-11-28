package helm

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type InstallArguments struct {
	Name       string            `yaml:"name"`
	Chart      string            `yaml:"chart"`
	Parameters map[string]string `yaml:"parameters"`
}

func (m *Mixin) Install() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}
	var args InstallArguments
	err = yaml.Unmarshal(payload, &args)
	if err != nil {
		return err
	}

	cmd := m.NewCommand("helm", "install", "--name", args.Name, args.Chart)
	cmd.Stdout = m.Out
	cmd.Stderr = m.Err

	prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
	if m.Debug {
		fmt.Fprintln(m.Out, prettyCmd)
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not execute command, %s: %s", prettyCmd, err)
	}

	return cmd.Wait()
}
