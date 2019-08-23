//go:generate packr2

package exec

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/deislabs/porter/pkg/context"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Exec is the logic behind the exec mixin
type Mixin struct {
	*context.Context

	Action Action

	schemas *packr.Box
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

	args := make([]string, len(step.Arguments), 1+len(step.Arguments)+len(step.Flags)*2)

	copy(args, step.Arguments)
	args = append(args, step.Flags.ToSlice()...)

	cmd := m.NewCommand(step.Command, args...)
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
