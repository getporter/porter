package exec

import (
	"github.com/deislabs/porter/pkg/exec/builder"
	yaml "gopkg.in/yaml.v2"
)

// ExecOptions represent the options for any exec command
type ExecuteOptions struct {
	File string
}

func (m *Mixin) loadAction(commandFile string) (*Action, error) {
	var action Action
	err := builder.LoadAction(m.Context, commandFile, func(contents []byte) (interface{}, error) {
		err := yaml.Unmarshal(contents, &action)
		return &action, err
	})
	return &action, err
}

func (m *Mixin) Execute(opts ExecuteOptions) error {
	action, err := m.loadAction(opts.File)
	if err != nil {
		return err
	}

	_, err = builder.ExecuteSingleStepAction(m.Context, action)
	return err
}
