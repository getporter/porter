package exec

import (
	"context"

	"get.porter.sh/porter/pkg/exec/builder"
	"get.porter.sh/porter/pkg/yaml"
)

// ExecOptions represent the options for any exec command
type ExecuteOptions struct {
	File string
}

func (m *Mixin) loadAction(ctx context.Context, commandFile string) (*Action, error) {
	var action Action
	err := builder.LoadAction(ctx, m.Config, commandFile, func(contents []byte) (interface{}, error) {
		err := yaml.Unmarshal(contents, &action)
		return &action, err
	})
	return &action, err
}

func (m *Mixin) Execute(ctx context.Context, opts ExecuteOptions) error {
	action, err := m.loadAction(ctx, opts.File)
	if err != nil {
		return err
	}

	_, err = builder.ExecuteSingleStepAction(ctx, m.Config, action)
	return err
}
