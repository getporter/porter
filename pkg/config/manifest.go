package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Manifest struct {
	Mixins  []string    `yaml:"mixins,omitempty"`
	Install []MixinStep `yaml:"install"`
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

type MixinStep struct {
	Description string                 `yaml:"description"`
	Data        map[string]interface{} `yaml:",inline"`
}

func (s MixinStep) Validate() error {
	if len(s.Data) == 0 {
		return errors.New("no mixin specified")
	}
	if len(s.Data) > 1 {
		return errors.New("more than one mixin specified")
	}
	return nil
}

func (s MixinStep) GetMixinType() (string, error) {
	var mixinName string
	for k := range s.Data {
		mixinName = k
	}
	return mixinName, nil
}

func (s MixinStep) GetMixinData() (string, error) {
	var mixinData []byte
	for _, data := range s.Data {
		mixinData, _ = yaml.Marshal(data)
	}
	return string(mixinData), nil
}
