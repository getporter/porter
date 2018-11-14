package helm

import (
	"gopkg.in/yaml.v2"
)

type InstallArguments struct {
	Name       string            `yaml:"name"`
	Chart      string            `yaml:"chart"`
	Parameters map[string]string `yaml:"parameters"`
}

func (m *Mixin) Install() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}
	var args InstallArguments
	err = yaml.Unmarshal(payload, &args)
	if err != nil {
		return err
	}
	// TODO: implement
	return nil
}
