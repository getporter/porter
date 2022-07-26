package migrations

import (
	"time"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/schema"
)

// SourceClaim represents the file format of claim documents from v0.38
type SourceClaim struct {
	// SchemaVersion is the version of the claim schema.
	SchemaVersion schema.Version `json:"schemaVersion"`

	// Id of the claim.
	ID string `json:"id"`

	// Installation name.
	Installation string `json:"installation"`

	// Revision of the installation.
	Revision string `json:"revision"`

	// Created timestamp of the claim.
	Created time.Time `json:"created"`

	// Action executed against the installation.
	Action string `json:"action"`

	// Bundle is the definition of the bundle.
	Bundle bundle.Bundle `json:"bundle"`

	// BundleReference is the canonical reference to the bundle used in the action.
	BundleReference string `json:"bundleReference,omitempty"`

	// Parameters are the key/value pairs that were passed in during the operation.
	Parameters map[string]interface{} `json:"parameters,omitempty"`

	// Custom extension data applicable to a given runtime.
	Custom interface{} `json:"custom,omitempty"`
}
