package config

import (
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Manifest struct {
	Name       string                `yaml:"image,omitempty"`
	Version    string                `yaml:"version,omitempty"`
	Image      string                `yaml:"invocationImage,omitempty"`
	Mixins     []string              `yaml:"mixins,omitempty"`
	Install    Steps                 `yaml:"install"`
	Parameters []ParameterDefinition `yaml:"parameters,omitempty"`
}

// ParameterDefinition defines a single parameter for a CNAB bundle
type ParameterDefinition struct {
	Name          string            `yaml:"name"`
	DataType      string            `yaml:"type"`
	DefaultValue  interface{}       `yaml:"default,omitempty"`
	AllowedValues []interface{}     `yaml:"allowed,omitempty"`
	Required      bool              `yaml:"required"`
	MinValue      *int              `yaml:"minValue,omitempty"`
	MaxValue      *int              `yaml:"maxValue,omitempty"`
	MinLength     *int              `yaml:"minLength,omitempty"`
	MaxLength     *int              `yaml:"maxLength,omitempty"`
	Metadata      ParameterMetadata `yaml:"metadata,omitempty"`
	Destination   *Location         `yaml:"destination,omitempty"`
}

type Location struct {
	Path                string `yaml:"path"`
	EnvironmentVariable string `yaml:"env"`
}

// ParameterMetadata contains metadata for a parameter definition.
type ParameterMetadata struct {
	Description string `yaml:"description,omitempty"`
}

func (c *Config) LoadManifest(file string) error {
	data, err := c.FileSystem.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "could not read manifest at %q", file)
	}

	m := &Manifest{}
	err = yaml.Unmarshal(data, m)
	if err != nil {
		return errors.Wrapf(err, "could not parse manifest yaml in %q", file)
	}
	c.Manifest = m
	return nil
}

func (m *Manifest) Validate() error {
	if len(m.Mixins) == 0 {
		return errors.New("no mixins declared")
	}
	if m.Install == nil {
		return errors.New("no install action defined")
	}
	return m.Install.Validate(m)
}

func (m *Manifest) GetSteps(action Action) (Steps, error) {
	var steps Steps
	switch action {
	case ActionInstall:
		steps = m.Install
	}

	if len(steps) == 0 {
		return nil, errors.Errorf("unsupported action: %q", action)
	}

	return steps, nil
}

type Steps []*Step

func (s Steps) Validate(m *Manifest) error {
	for _, step := range s {
		err := step.Validate(m)
		if err != nil {
			return err
		}
	}
	return nil
}

type Step struct {
	Description string                 `yaml:"description"`
	Data        map[string]interface{} `yaml:",inline"`

	runner *mixin.Runner
}

func (s *Step) Validate(m *Manifest) error {
	if len(s.Data) == 0 {
		return errors.New("no mixin specified")
	}
	if len(s.Data) > 1 {
		return errors.New("more than one mixin specified")
	}

	mixinDeclared := false
	mixinType := s.GetMixinName()
	for _, mixin := range m.Mixins {
		if mixin == mixinType {
			mixinDeclared = true
			break
		}
	}
	if !mixinDeclared {
		return errors.Errorf("mixin (%s) was not declared", mixinType)
	}

	return nil
}

func (s *Step) GetMixinName() string {
	var mixinName string
	for k := range s.Data {
		mixinName = k
	}
	return mixinName
}

func (s *Step) GetMixinData() string {
	var mixinData []byte
	for _, data := range s.Data {
		mixinData, _ = yaml.Marshal(data)
	}
	return string(mixinData)
}
