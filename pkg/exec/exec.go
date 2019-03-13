//go:generate packr2

package exec

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/context"
	"github.com/gobuffalo/packr/v2"
	yaml "gopkg.in/yaml.v2"
)

// Exec is the logic behind the exec mixin
type Mixin struct {
	*context.Context

	Action Action

	schemas *packr.Box
}

type Action struct {
	Steps []Step // using UnmarshalYAML so that we don't need a custom type per action
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - exec: ...
// and puts the steps into the Action.Steps field
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	actionMap := map[interface{}][]interface{}{}
	err := unmarshal(&actionMap)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal yaml into an action map of exec steps")
	}

	for _, stepMaps := range actionMap {
		b, err := yaml.Marshal(stepMaps)
		if err != nil {
			return err
		}

		var steps []Step
		err = yaml.Unmarshal(b, &steps)
		if err != nil {
			return err
		}

		a.Steps = append(a.Steps, steps...)
	}

	return nil
}

type Step struct {
	Instruction `yaml:"exec"`
}

type Instruction struct {
	Description string   `yaml:"description"`
	Command     string   `yaml:"command"`
	Arguments   []string `yaml:"arguments"`
}

// New exec mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context: context.New(),
		schemas: NewSchemaBox(),
	}
}

func NewSchemaBox() *packr.Box {
	return packr.New("github.com/deislabs/porter/pkg/exec/schema", "./schema")
}

func (m *Mixin) loadAction(commandFile string) error {
	contents, err := m.getCommandFile(commandFile, m.Out)
	if err != nil {
		source := "STDIN"
		if commandFile == "" {
			source = commandFile
		}
		return errors.Wrapf(err, "could not load input from %s", source)
	}

	err = yaml.Unmarshal(contents, &m.Action)
	if m.Debug {
		fmt.Fprintf(m.Err, "DEBUG Parsed Input:\n%#v\n", m.Action)
	}
	return errors.Wrapf(err, "could unmarshal input:\n %s", string(contents))
}

func (m *Mixin) Execute() error {
	if len(m.Action.Steps) != 1 {
		return errors.Errorf("expected a single step, but got %d", len(m.Action.Steps))
	}
	step := m.Action.Steps[0]

	cmd := m.NewCommand(step.Command, step.Arguments...)
	cmd.Stdout = m.Out
	cmd.Stderr = m.Err

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start...%s", err)
	}

	return cmd.Wait()
}

func (m *Mixin) getCommandFile(commandFile string, w io.Writer) ([]byte, error) {
	if commandFile == "" {
		reader := bufio.NewReader(m.In)
		return ioutil.ReadAll(reader)
	}
	return m.FileSystem.ReadFile(commandFile)
}
