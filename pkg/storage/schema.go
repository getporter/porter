package storage

import (
	"github.com/cnabio/cnab-go/schema"
)

var _ Document = Schema{}

const (
	// InstallationSchemaVersion represents the version associated with the schema
	// for all installation documents: installations, runs, results and outputs.
	InstallationSchemaVersion = schema.Version("1.0.2")

	// CredentialSetSchemaVersion represents the version associated with the schema
	// credential set documents.
	CredentialSetSchemaVersion = schema.Version("1.0.1")

	// ParameterSetSchemaVersion represents the version associated with the schema
	// for parameter set documents.
	ParameterSetSchemaVersion = schema.Version("1.0.1")
)

type Schema struct {
	ID string `json:"_id"`

	// Installations is the schema for the installation documents.
	Installations schema.Version `json:"installations"`

	// Credentials is the schema for the credential spec documents.
	Credentials schema.Version `json:"credentials"`

	// Parameters is the schema for the parameter spec documents.
	Parameters schema.Version `json:"parameters"`
}

// NewSchema creates a schema document with the currently supported version for all subsystems.
func NewSchema() Schema {
	return Schema{
		ID:            "schema",
		Installations: InstallationSchemaVersion,
		Credentials:   CredentialSetSchemaVersion,
		Parameters:    ParameterSetSchemaVersion,
	}
}

func (s Schema) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"_id": "schema"}
}

func (s Schema) IsOutOfDate() bool {
	return s.ShouldMigrateInstallations() || s.ShouldMigrateCredentialSets() || s.ShouldMigrateParameterSets()
}

func (s Schema) ShouldMigrateInstallations() bool {
	return s.Installations != InstallationSchemaVersion
}

func (s Schema) ShouldMigrateCredentialSets() bool {
	return s.Credentials != CredentialSetSchemaVersion
}

func (s Schema) ShouldMigrateParameterSets() bool {
	return s.Parameters != ParameterSetSchemaVersion
}
