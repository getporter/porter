package parameters

import (
	"fmt"
	"time"

	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/secrets"
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
	Parameters []ParameterStrategy `json:"parameters" yaml:"parameters"`
}

// TODO: is this needed?  Clone of cnab-go's credentials.ResolveCredentials(...)
//
// Resolve looks up the parameters as described in Source, then copies
// the resulting value into the Value field of each parameter strategy.
//
// The typical workflow for working with a parameter set is:
//
//	- Load the set
//	- Validate the parameters against a spec
//	- Resolve the parameters
//	- Expand them into bundle values
func (p *ParameterSet) Resolve(s secrets.Store) (credentials.Set, error) {
	l := len(p.Parameters)
	res := make(map[string]string, l)
	for i := 0; i < l; i++ {
		param := p.Parameters[i]
		val, err := s.Resolve(param.Source.Key, param.Source.Value)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %v", p.Parameters[i].Name, err)
		}
		param.Value = val
		res[p.Parameters[i].Name] = param.Value
	}
	return res, nil
}

// ParameterStrategy represents a strategy to resolve a value source
// for a parameter
type ParameterStrategy struct {
	// Name is the name of the parameter.
	// Name is used to match a parameter strategy to a bundle's parameter.
	Name string `json:"name" yaml:"name"`
	// Source is the location of the parameter.
	// During resolution, the source will be loaded, and the result temporarily placed
	// into Value.
	Source credentials.Source `json:"source,omitempty" yaml:"source,omitempty"`
	// Value holds the parameter value.
	// When a parameter is loaded, it is loaded into this field. In all
	// other cases, it is empty. This field is omitted during serialization.
	Value string `json:"-" yaml:"-"`
}
