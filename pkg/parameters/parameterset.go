package parameters

import (
	"time"

	"github.com/cnabio/cnab-go/valuesource"
)

// ParameterSet represents a collection of parameters and their
// sources/strategies for value resolution
type ParameterSet struct {
	// Name is the name of the parameter set.
	Name string `json:"name" yaml:"name"`
	// Created timestamp of the parameter set.
	Created time.Time `json:"created" yaml:"created"`
	// Modified timestamp of the parameter set.
	Modified time.Time `json:"modified" yaml:"modified"`
	// Parameters is a list of parameter specs.
	Parameters []valuesource.Strategy `json:"parameters" yaml:"parameters"`
}
