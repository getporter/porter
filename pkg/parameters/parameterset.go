package parameters

import (
	"fmt"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/schema"
	"github.com/pkg/errors"
)

const INTERNAL_PARAMETERER_SET = "internal-parameter-set"

var _ storage.Document = ParameterSet{}

// ParameterSet represents a collection of parameters and their
// sources/strategies for value resolution
type ParameterSet struct {
	ParameterSetSpec `yaml:",inline"`
	Status           ParameterSetStatus `json:"status" yaml:"status" toml:"status"`
}

// ParameterSetSpec represents the set of user-modifiable fields on a ParameterSet.
type ParameterSetSpec struct {
	// SchemaVersion is the version of the parameter-set schema.
	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`

	// Namespace to which the credential set is scoped.
	Namespace string `json:"namespace" yaml:"namespace" toml:"namespace"`

	// Name is the name of the parameter set.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Labels applied to the parameter set.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" toml:"labels,omitempty"`

	// Parameters is a list of parameter specs.
	Parameters []secrets.Strategy `json:"parameters" yaml:"parameters" toml:"parameters"`
}

// ParameterSetStatus contains additional status metadata that has been set by Porter.
type ParameterSetStatus struct {
	// Created timestamp of the parameter set.
	Created time.Time `json:"created" yaml:"created" toml:"created"`

	// Modified timestamp of the parameter set.
	Modified time.Time `json:"modified" yaml:"modified" toml:"modified"`
}

// NewParameterSet creates a new ParameterSet with the required fields initialized.
func NewParameterSet(namespace string, name string, params ...secrets.Strategy) ParameterSet {
	now := time.Now()
	ps := ParameterSet{
		ParameterSetSpec: ParameterSetSpec{
			SchemaVersion: SchemaVersion,
			Namespace:     namespace,
			Name:          name,
			Parameters:    params,
		},
		Status: ParameterSetStatus{
			Created:  now,
			Modified: now,
		},
	}

	return ps
}

// NewParameterSet creates a new ParameterSet with the required fields initialized.
func NewInternalParameterSet(namespace string, name string, params ...secrets.Strategy) ParameterSet {
	return NewParameterSet(namespace, INTERNAL_PARAMETERER_SET+"-"+name, params...)
}

func (s ParameterSet) IsInternalParameterSet() bool {
	return strings.Contains(s.Name, INTERNAL_PARAMETERER_SET)
}

func (s ParameterSet) DefaultDocumentFilter() interface{} {
	return map[string]interface{}{"namespace": s.Namespace, "name": s.Name}
}

func (s ParameterSet) Validate() error {
	if SchemaVersion != s.SchemaVersion {
		if s.SchemaVersion == "" {
			s.SchemaVersion = "(none)"
		}
		return errors.Errorf("invalid schemaVersion provided: %s. This version of Porter is compatible with %s.", s.SchemaVersion, SchemaVersion)
	}
	return nil
}

func (s ParameterSet) String() string {
	if s.Namespace != "" {
		return fmt.Sprintf("%s/%s", s.Namespace, s.Name)
	}

	return s.Name
}

func (s ParameterSet) Resolve(resolver Provider, bun cnab.ExtendedBundle) (map[string]interface{}, error) {
	params, err := resolver.ResolveAll(s)
	if err != nil {
		return nil, err
	}

	resolved := make(map[string]interface{})
	for _, param := range params {
		paramValue, err := bun.ConvertParameterValue(param.Name, param.Value)
		if err != nil {
			paramValue = param.Value
		}

		resolved[param.Name] = paramValue

	}
	return resolved, nil
}
