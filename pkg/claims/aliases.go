package claims

import (
	cnabclaims "github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/schema"
)

// Alias common cnab values in this package so that we don't have imports from
// this package and the cnab-go package which gets super confusing now that we
// are declaring document types in this package.

// CNABSchemaVersion is the schemaVersion value for CNAB documents such as claims.
func CNABSchemaVersion() schema.Version {
	return cnabclaims.GetDefaultSchemaVersion()
}

type OutputMetadata = cnabclaims.OutputMetadata
