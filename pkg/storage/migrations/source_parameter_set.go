package migrations

import (
	"time"

	"github.com/cnabio/cnab-go/schema"
	"github.com/cnabio/cnab-go/valuesource"
)

// SourceParameterSet represents the file format of credential set documents from v0.38
type SourceParameterSet struct {
	// SchemaVersion is the version of the parameter-set schema.
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
