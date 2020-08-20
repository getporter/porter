package parameters

import (
	"time"

	"github.com/cnabio/cnab-go/schema"
	"github.com/cnabio/cnab-go/valuesource"
)

const (
	// DefaultSchemaVersion is the default SchemaVersion value
	// set on new CredentialSet instances, and is the semver portion
	// of CNABSpecVersion.
	DefaultSchemaVersion = schema.Version("1.0.0-DRAFT+TODO")

	// CNABSpecVersion represents the CNAB Spec version of the Credentials
	// that this library implements
	// This value is prefixed with e.g. `cnab-credentials-` so isn't itself valid semver.
	CNABSpecVersion string = "cnab-parametersets-" + string(DefaultSchemaVersion)
)

// ParameterSet represents a collection of parameters and their
// sources/strategies for value resolution
type ParameterSet struct {
	// SchemaVersion is the version of the paramete-set schema.
	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion"`
	// Name is the name of the parameter set.
	Name string `json:"name" yaml:"name"`
	// Created timestamp of the parameter set.
	Created time.Time `json:"created" yaml:"created"`
	// Modified timestamp of the parameter set.
	Modified time.Time `json:"modified" yaml:"modified"`
	// Parameters is a list of parameter specs.
	Parameters []valuesource.Strategy `json:"parameters" yaml:"parameters"`
}

// NewParameterSet creates a new ParameterSet with the required fields initialized.
func NewParameterSet(name string, params ...valuesource.Strategy) ParameterSet {
	now := time.Now()
	ps := ParameterSet{
		SchemaVersion: DefaultSchemaVersion,
		Name:          name,
		Created:       now,
		Modified:      now,
		Parameters:    params,
	}

	return ps
}
