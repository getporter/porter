package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Manifest struct {
	Mixins  []string `yaml:"mixins,omitempty"`
	Install Action   `yaml:"install"`
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
	if m.Install == nil {
		return errors.New("no install action defined")
	}
	return m.Install.Validate(m)
}

type Action []MixinStep

func (a Action) Validate(m *Manifest) error {
	for _, step := range a {
		err := step.Validate(m)
		if err != nil {
			return err
		}
	}
	return nil
}

type MixinStep struct {
	Description string                 `yaml:"description"`
	Data        map[string]interface{} `yaml:",inline"`
}

func (s MixinStep) Validate(m *Manifest) error {
	if len(s.Data) == 0 {
		return errors.New("no mixin specified")
	}
	if len(s.Data) > 1 {
		return errors.New("more than one mixin specified")
	}

	mixinDeclared := false
	mixinType := s.GetMixinType()
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

func (s MixinStep) GetMixinType() string {
	var mixinName string
	for k := range s.Data {
		mixinName = k
	}
	return mixinName
}

func (s MixinStep) GetMixinData() string {
	var mixinData []byte
	for _, data := range s.Data {
		mixinData, _ = yaml.Marshal(data)
	}
	return string(mixinData)
}
