package parameters

import (
	"time"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/schema"
)

const (
	// DefaultSchemaVersion is the default SchemaVersion value
	// set on new ParameterSet instances, and is the semver portion
	// of CNABSpecVersion.
	DefaultSchemaVersion = schema.Version("1.0.0-DRAFT+TODO")

	// CNABSpecVersion represents the CNAB Spec version of the Parameters
	// that this library implements.
	// This value is prefixed with e.g. `cnab-parametersets-` so isn't itself valid semver.
	CNABSpecVersion string = "cnab-parametersets-" + string(DefaultSchemaVersion)
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

	// Parameters is a list of parameter specs.
	Parameters []secrets.Strategy `json:"parameters" yaml:"parameters" toml:"parameters"`
}

// NewParameterSet creates a new ParameterSet with the required fields initialized.
func NewParameterSet(namespace string, name string, params ...secrets.Strategy) ParameterSet {
	now := time.Now()
	ps := ParameterSet{
		SchemaVersion: DefaultSchemaVersion,
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
