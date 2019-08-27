package exec

import (
	"github.com/deislabs/porter/pkg/exec/builder"
)

type Action struct {
	Steps []Step // using UnmarshalYAML so that we don't need a custom type per action
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - exec: ...
// and puts the steps into the Action.Steps field
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var steps []Step
	results, err := builder.UnmarshalAction(unmarshal, &steps)
	if err != nil {
		return err
	}

	for _, result := range results {
		step := result.(*[]Step)
		a.Steps = append(a.Steps, *step...)
	}
	return nil
}

type Step struct {
	Instruction `yaml:"exec"`
}

type Instruction struct {
	Description string        `yaml:"description"`
	Command     string        `yaml:"command"`
	Arguments   []string      `yaml:"arguments,omitempty"`
	Flags       builder.Flags `yaml:"flags,omitempty"`
}
