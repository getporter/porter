package exec

import (
	"github.com/deislabs/porter/pkg/exec/builder"
)

var _ builder.ExecutableAction = Action{}

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

func (a Action) GetSteps() []builder.ExecutableStep {
	steps := make([]builder.ExecutableStep, len(a.Steps))
	for i := range a.Steps {
		steps[i] = a.Steps[i]
	}

	return steps
}

var _ builder.ExecutableStep = Step{}
var _ builder.HasCustomDashes = Step{}
var _ builder.StepWithOutputs = Step{}

type Step struct {
	Instruction `yaml:"exec"`
}

type Instruction struct {
	Description string        `yaml:"description"`
	Command     string        `yaml:"command"`
	Arguments   []string      `yaml:"arguments,omitempty"`
	Flags       builder.Flags `yaml:"flags,omitempty"`
	Outputs     []Output      `yaml:"outputs,omitempty"`
}

func (s Step) GetCommand() string {
	return s.Command
}

func (s Step) GetArguments() []string {
	return s.Arguments
}

func (s Step) GetFlags() builder.Flags {
	return s.Flags
}

func (s Step) GetDashes() builder.Dashes {
	return builder.DefaultFlagDashes
}

func (s Step) GetOutputs() []builder.Output {
	outputs := make([]builder.Output, len(s.Outputs))
	for i := range s.Outputs {
		outputs[i] = s.Outputs[i]
	}
	return outputs
}

var _ builder.OutputRegex = Output{}
var _ builder.OutputFile = Output{}
var _ builder.OutputJsonPath = Output{}

type Output struct {
	Name     string `yaml:"name"`
	FilePath string `yaml:"path,omitempty"`
	JsonPath string `yaml:"jsonPath,omitempty"`
	Regex    string `yaml:"regex,omitempty"`
}

func (o Output) GetName() string {
	return o.Name
}

func (o Output) GetFilePath() string {
	return o.FilePath
}

func (o Output) GetJsonPath() string {
	return o.JsonPath
}

func (o Output) GetRegex() string {
	return o.Regex
}
