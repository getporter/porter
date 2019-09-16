package mixin

import (
	"github.com/deislabs/porter/pkg/config"
)

type BuildInput struct {
	Config  interface{}             `yaml:"config,omitempty"`
	Actions map[string]config.Steps `yaml:"actions"`
}
