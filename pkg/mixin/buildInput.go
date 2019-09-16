package mixin

import (
	"github.com/deislabs/porter/pkg/config"
)

type BuildInput struct {
	Config  map[string]interface{}  `yaml:"config"`
	Actions map[string]config.Steps `yaml:"actions"`
}
