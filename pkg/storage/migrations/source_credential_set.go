package migrations

import (
	"time"

	"github.com/cnabio/cnab-go/schema"
	"github.com/cnabio/cnab-go/valuesource"
)

// SourceCredentialSet represents the file format of credential set documents from v0.38
type SourceCredentialSet struct {
	// SchemaVersion is the version of the credential-set schema.
	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion"`

	// Name is the name of the credentialset.
	Name string `json:"name" yaml:"name"`

	// Created timestamp of the credentialset.
	Created time.Time `json:"created" yaml:"created"`

	// Modified timestamp of the credentialset.
	Modified time.Time `json:"modified" yaml:"modified"`

	// Credentials is a list of credential specs.
	Credentials []valuesource.Strategy `json:"credentials" yaml:"credentials"`
}
