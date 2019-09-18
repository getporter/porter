package mixin

type BuildInput struct {
	Config  interface{}            `yaml:"config,omitempty"`
	Actions map[string]interface{} `yaml:"actions"`
}
