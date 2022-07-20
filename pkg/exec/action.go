package exec

import (
	"get.porter.sh/porter/pkg/exec/builder"
)

var _ builder.ExecutableAction = Action{}
var _ builder.BuildableAction = Action{}

type Action struct {
	Name  string
	Steps []Step // using UnmarshalYAML so that we don't need a custom type per action
}

// MarshalYAML converts the action back to a YAML representation
// install:
//   exec:
//     ...
//   helm3:
//     ...
func (a Action) MarshalYAML() (interface{}, error) {
	return map[string]interface{}{a.Name: a.Steps}, nil
}

// MakeSteps builds a slice of Steps for data to be unmarshaled into.
func (a Action) MakeSteps() interface{} {
	return &[]Step{}
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - exec: ...
// and puts the steps into the Action.Steps field
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	results, err := builder.UnmarshalAction(unmarshal, a)
	if err != nil {
		return err
	}

	for actionName, action := range results {
		a.Name = actionName
		for _, result := range action {
			step := result.(*[]Step)
			a.Steps = append(a.Steps, *step...)
		}
		break // There is only 1 action
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

// Actions is a set of actions, and the steps, passed from Porter.
type Actions []Action

// UnmarshalYAML takes chunks of a porter.yaml file associated with this mixin
// and populates it on the current action set.
// install:
//   exec:
//     ...
//   exec:
//     ...
// upgrade:
//   exec:
//     ...
func (a *Actions) UnmarshalYAML(unmarshal func(interface{}) error) error {
	results, err := builder.UnmarshalAction(unmarshal, Action{})
	if err != nil {
		return err
	}

	for actionName, action := range results {
		for _, result := range action {
			s := result.(*[]Step)
			*a = append(*a, Action{
				Name:  actionName,
				Steps: *s,
			})
		}
	}
	return nil
}

var (
	_ builder.HasOrderedArguments = Step{}
	_ builder.ExecutableStep      = Step{}
	_ builder.StepWithOutputs     = Step{}
	_ builder.HasEnvironmentVars  = Step{}
)

type Step struct {
	Instruction `yaml:"exec"`
}

type Instruction struct {
	Description     string            `yaml:"description"`
	Command         string            `yaml:"command"`
	WorkingDir      string            `yaml:"dir,omitempty"`
	Arguments       []string          `yaml:"arguments,omitempty"`
	SuffixArguments []string          `yaml:"suffix-arguments,omitempty"`
	Flags           builder.Flags     `yaml:"flags,omitempty"`
	EnvironmentVars map[string]string `yaml:"envs,omitempty"`
	Outputs         []Output          `yaml:"outputs,omitempty"`
	SuppressOutput  bool              `yaml:"suppress-output,omitempty"`

	// Allow the user to ignore some errors
	builder.IgnoreErrorHandler `yaml:"ignoreError,omitempty"`
}

func (s Step) GetCommand() string {
	return s.Command
}

func (s Step) GetArguments() []string {
	return s.Arguments
}

func (s Step) GetSuffixArguments() []string {
	return s.SuffixArguments
}

func (s Step) GetFlags() builder.Flags {
	return s.Flags
}

func (s Step) GetEnvironmentVars() map[string]string {
	return s.EnvironmentVars
}

func (s Step) SuppressesOutput() bool {
	return s.SuppressOutput
}

func (s Step) GetOutputs() []builder.Output {
	outputs := make([]builder.Output, len(s.Outputs))
	for i := range s.Outputs {
		outputs[i] = s.Outputs[i]
	}
	return outputs
}

func (s Step) GetWorkingDir() string {
	return s.WorkingDir
}

var _ builder.OutputRegex = Output{}
var _ builder.OutputFile = Output{}
var _ builder.OutputJsonPath = Output{}

type Output struct {
	Name      string `yaml:"name"`
	FilePath  string `yaml:"path,omitempty"`
	JsonPath  string `yaml:"jsonPath,omitempty"`
	Regex     string `yaml:"regex,omitempty"`
	Sensitive bool   `yaml:"sensitive,omitempty"`
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

func (o Output) GetMetadata() builder.OutputMetadata {
	return builder.OutputMetadata{
		Sensitive: o.Sensitive,
	}
}
