package parameters

import (
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/schema"
	"github.com/pkg/errors"
)

var _ storage.Document = ParameterSet{}

// ParameterSet represents a collection of parameters and their
// sources/strategies for value resolution
type ParameterSet struct {
	// SchemaVersion is the version of the parameter-set schema.
	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`

	// Namespace to which the credential set is scoped.
	Namespace string `json:"namespace" yaml:"namespace" toml:"namespace"`

	// Name is the name of the parameter set.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Created timestamp of the parameter set.
	Created time.Time `json:"created" yaml:"created" toml:"created"`

	// Modified timestamp of the parameter set.
	Modified time.Time `json:"modified" yaml:"modified" toml:"modified"`

	// Labels applied to the parameter set.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" toml:"labels,omitempty"`

	// Parameters is a list of parameter specs.
	Parameters []secrets.Strategy `json:"parameters" yaml:"parameters" toml:"parameters"`
}

// NewParameterSet creates a new ParameterSet with the required fields initialized.
func NewParameterSet(namespace string, name string, params ...secrets.Strategy) ParameterSet {
	now := time.Now()
	ps := ParameterSet{
		SchemaVersion: SchemaVersion,
		Namespace:     namespace,
		Name:          name,
		Created:       now,
		Modified:      now,
		Parameters:    params,
	}

	return ps
}

func (s ParameterSet) DefaultDocumentFilter() interface{} {
	return map[string]interface{}{"namespace": s.Namespace, "name": s.Name}
}

func (s ParameterSet) Validate() error {
	if SchemaVersion != s.SchemaVersion {
		return errors.Errorf("invalid schemaVersion provided: %s. This version of Porter is compatible with %s.", s.SchemaVersion, SchemaVersion)
	}
	return nil
}

func (s ParameterSet) String() string {
	return fmt.Sprintf("%s/%s", s.Namespace, s.Name)
}
