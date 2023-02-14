package storage

import (
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/schema"
	"github.com/Masterminds/semver/v3"
)

var _ Document = Schema{}

const (
	// SchemaTypeCredentialSet is the default schemaType value for CredentialSet resources
	SchemaTypeCredentialSet = "CredentialSet"

	// SchemaTypeInstallation is the default schemaType value for Installation resources
	SchemaTypeInstallation = "Installation"

	// SchemaTypeParameterSet is the default schemaType value for ParameterSet resources
	SchemaTypeParameterSet = "ParameterSet"

	// DefaultCredentialSetSchemaVersion represents the version associated with the schema
	// credential set documents.
	DefaultCredentialSetSchemaVersion = cnab.SchemaVersion("1.0.1")

	// DefaultInstallationSchemaVersion represents the version associated with the schema
	// for all installation documents: installations, runs, results and outputs.
	DefaultInstallationSchemaVersion = cnab.SchemaVersion("1.0.2")

	// DefaultParameterSetSchemaVersion represents the version associated with the schema
	//	// for parameter set documents.
	DefaultParameterSetSchemaVersion = cnab.SchemaVersion("1.0.1")
)

var (
	// DefaultCredentialSetSemverSchemaVersion is the semver representation of the DefaultCredentialSetSchemaVersion  that is suitable for doing semver comparisons.
	DefaultCredentialSetSemverSchemaVersion = semver.MustParse(string(DefaultCredentialSetSchemaVersion))

	// DefaultInstallationSemverSchemaVersion is the semver representation of the DefaultInstallationSchemaVersion that is suitable for doing semver comparisons.
	DefaultInstallationSemverSchemaVersion = semver.MustParse(string(DefaultInstallationSchemaVersion))

	// DefaultParameterSetSemverSchemaVersion is the semver representation of the DefaultParameterSetSchemaVersion  that is suitable for doing semver comparisons.
	DefaultParameterSetSemverSchemaVersion = semver.MustParse(string(DefaultParameterSetSchemaVersion))

	// SupportedCredentialSetSchemaVersions represents the set of allowed schema versions for CredentialSet documents.
	SupportedCredentialSetSchemaVersions = schema.MustParseConstraint("1.0.1")

	// SupportedInstallationSchemaVersions represents the set of allowed schema versions for Installation documents.
	SupportedInstallationSchemaVersions = schema.MustParseConstraint("1.0.2")

	// SupportedParameterSetSchemaVersions represents the set of allowed schema versions for ParameterSet documents.
	SupportedParameterSetSchemaVersions = schema.MustParseConstraint("1.0.1")
)

type Schema struct {
	ID string `json:"_id"`

	// Installations is the schema for the installation documents.
	Installations cnab.SchemaVersion `json:"installations"`

	// Credentials is the schema for the credential spec documents.
	Credentials cnab.SchemaVersion `json:"credentials"`

	// Parameters is the schema for the parameter spec documents.
	Parameters cnab.SchemaVersion `json:"parameters"`
}

// NewSchema creates a schema document with the currently supported version for all subsystems.
func NewSchema() Schema {
	return Schema{
		ID:            "schema",
		Installations: DefaultInstallationSchemaVersion,
		Credentials:   DefaultCredentialSetSchemaVersion,
		Parameters:    DefaultParameterSetSchemaVersion,
	}
}

func (s Schema) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"_id": "schema"}
}

func (s Schema) IsOutOfDate() bool {
	return s.ShouldMigrateInstallations() || s.ShouldMigrateCredentialSets() || s.ShouldMigrateParameterSets()
}

// ShouldMigrateInstallations checks if the minimum version of the installation resources in the database is unsupported and requires a migration to work with this version of Porter.
// Since Porter can support a range of resource versions, this means that the db may have multiple representations of a resource in the database, and will migrate them to the latest support schema version on an as-needed basis.
func (s Schema) ShouldMigrateInstallations() bool {
	// Determine if the minimum installation version in the db is completely unsupported by this version of Porter
	warnOnly, err := schema.ValidateSchemaVersion(schema.CheckStrategyExact, SupportedInstallationSchemaVersions, string(s.Installations), DefaultInstallationSemverSchemaVersion)
	return !warnOnly && err != nil
}

// ShouldMigrateCredentialSets checks if the minimum version of the Credential set resources in the database is unsupported and requires a migration to work with this version of Porter.
// Since Porter can support a range of resource versions, this means that the db may have multiple representations of a resource in the database, and will migrate them to the latest support schema version on an as-needed basis.
func (s Schema) ShouldMigrateCredentialSets() bool {
	// Determine if the minimum Credential set version in the db is completely unsupported by this version of Porter
	warnOnly, err := schema.ValidateSchemaVersion(schema.CheckStrategyExact, SupportedCredentialSetSchemaVersions, string(s.Credentials), DefaultCredentialSetSemverSchemaVersion)
	return !warnOnly && err != nil
}

// ShouldMigrateParameterSets checks if the minimum version of the parameter set resources in the database is unsupported and requires a migration to work with this version of Porter.
// Since Porter can support a range of resource versions, this means that the db may have multiple representations of a resource in the database, and will migrate them to the latest support schema version on an as-needed basis.
func (s Schema) ShouldMigrateParameterSets() bool {
	// Determine if the minimum paramter set version in the db is completely unsupported by this version of Porter
	warnOnly, err := schema.ValidateSchemaVersion(schema.CheckStrategyExact, SupportedParameterSetSchemaVersions, string(s.Parameters), DefaultParameterSetSemverSchemaVersion)
	return !warnOnly && err != nil
}
