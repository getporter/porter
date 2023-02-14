package storage

import (
	"fmt"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/secrets"
)

const (
	// SchemaTypeParameterSet is the default schemaType value for ParameterSet resources
	SchemaTypeParameterSet = "ParameterSet"

	INTERNAL_PARAMETERER_SET = "internal-parameter-set"
)

var _ Document = ParameterSet{}

// ParameterSet represents a collection of parameters and their
// sources/strategies for value resolution
type ParameterSet struct {
	ParameterSetSpec `yaml:",inline"`
	Status           ParameterSetStatus `json:"status" yaml:"status" toml:"status"`
}

// ParameterSetSpec represents the set of user-modifiable fields on a ParameterSet.
type ParameterSetSpec struct {
	// SchemaType helps when we export the definition so editors can detect the type of document, it's not used by porter.
	SchemaType string `json:"schemaType,omitempty" yaml:"schemaType,omitempty"`

	// SchemaVersion is the version of the parameter-set schema.
	SchemaVersion cnab.SchemaVersion `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`

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
			SchemaType:    SchemaTypeParameterSet,
			SchemaVersion: ParameterSetSchemaVersion,
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

// NewInternalParameterSet creates a new internal ParameterSet with the required fields initialized.
func NewInternalParameterSet(namespace string, name string, params ...secrets.Strategy) ParameterSet {
	return NewParameterSet(namespace, INTERNAL_PARAMETERER_SET+"-"+name, params...)
}

func (s ParameterSet) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"namespace": s.Namespace, "name": s.Name}
}

func (s *ParameterSet) Validate() error {
	if s.SchemaType == "" {
		// Default the schema type before importing into the database if it's not set already
		// SchemaType isn't really used by our code, it's a type hint for editors, but this will ensure we are consistent in our persisted documents
		s.SchemaType = SchemaTypeParameterSet
	} else if !strings.EqualFold(s.SchemaType, SchemaTypeParameterSet) {
		return fmt.Errorf("invalid schemaType %s, expected %s", s.SchemaType, SchemaTypeParameterSet)
	}

	if ParameterSetSchemaVersion != s.SchemaVersion {
		if s.SchemaVersion == "" {
			s.SchemaVersion = "(none)"
		}
		return fmt.Errorf("invalid schemaVersion provided: %s. This version of Porter is compatible with %s.", s.SchemaVersion, ParameterSetSchemaVersion)
	}
	return nil
}

func (s ParameterSet) String() string {
	return fmt.Sprintf("%s/%s", s.Namespace, s.Name)
}
