package helm

import (
	"bufio"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type InstallStep struct {
	Description string           `yaml:"description"`
	Outputs     []HelmOutput     `yaml:"outputs"`
	Arguments   InstallArguments `yaml:"helm"`
}

type HelmOutput struct {
	Name   string `yaml:"name"`
	Secret string `yaml:"secret"`
	Key    string `yaml:"key"`
}

type InstallArguments struct {
	Namespace string            `yaml:"namespace"`
	Name      string            `yaml:"name"`
	Chart     string            `yaml:"chart"`
	Version   string            `yaml:"version"`
	Replace   bool              `yaml:"replace"`
	Set       map[string]string `yaml:"set"`
	Values    []string          `yaml:"values"`
}

func (m *Mixin) Install() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}
	var step InstallStep
	err = yaml.Unmarshal(payload, &step)
	if err != nil {
		return err
	}

	cmd := m.NewCommand("helm", "install", "--name", step.Arguments.Name, step.Arguments.Chart)

	if step.Arguments.Namespace != "" {
		cmd.Args = append(cmd.Args, "--namespace", step.Arguments.Namespace)
	}

	if step.Arguments.Version != "" {
		cmd.Args = append(cmd.Args, "--version", step.Arguments.Version)
	}

	if step.Arguments.Replace {
		cmd.Args = append(cmd.Args, "--replace")
	}

	for _, v := range step.Arguments.Values {
		cmd.Args = append(cmd.Args, "--values", v)
	}

	// sort the set consistently
	setKeys := make([]string, 0, len(step.Arguments.Set))
	for k := range step.Arguments.Set {
		setKeys = append(setKeys, k)
	}
	sort.Strings(setKeys)

	for _, k := range setKeys {
		cmd.Args = append(cmd.Args, "--set", fmt.Sprintf("%s=%s", k, step.Arguments.Set[k]))
	}

	cmd.Stdout = m.Out
	cmd.Stderr = m.Err

	prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
	fmt.Fprintln(m.Out, prettyCmd)

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not execute command, %s: %s", prettyCmd, err)
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}

	f, err := m.Context.NewOutput()
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	defer buf.Flush()

	for _, output := range step.Outputs {
		val, err := m.getSecret(step.Arguments.Namespace, output.Secret, output.Key)
		if err != nil {
			return err
		}
		l := fmt.Sprintf("%s=%s\n", output.Name, val)
		buf.Write([]byte(l))
	}
	return cmd.Wait()
}
