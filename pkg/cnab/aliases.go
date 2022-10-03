package cnab

import (
	"github.com/cnabio/cnab-go/bundle"
	cnabclaims "github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/schema"
)

// Alias common cnab values in this package so that we don't have imports from
// this package and the cnab-go package which gets super confusing now that we
// are declaring document types in this package.

const (
	ActionInstall   = cnabclaims.ActionInstall
	ActionUpgrade   = cnabclaims.ActionUpgrade
	ActionUninstall = cnabclaims.ActionUninstall
	ActionUnknown   = cnabclaims.ActionUnknown

	StatusSucceeded = cnabclaims.StatusSucceeded
	StatusCanceled  = cnabclaims.StatusCanceled
	StatusFailed    = cnabclaims.StatusFailed
	StatusRunning   = cnabclaims.StatusRunning
	StatusPending   = cnabclaims.StatusPending
	StatusUnknown   = cnabclaims.StatusUnknown

	OutputInvocationImageLogs = cnabclaims.OutputInvocationImageLogs
)

type Installation = cnabclaims.Installation
type Claim = cnabclaims.Claim
type Result = cnabclaims.Result
type Output = cnabclaims.Output
type Outputs = cnabclaims.Outputs
type OutputMetadata = cnabclaims.OutputMetadata

var NewULID = cnabclaims.MustNewULID

// BundleSchemaVersion is the schemaVersion value for CNAB bundle documents.
func BundleSchemaVersion() schema.Version {
	return bundle.GetDefaultSchemaVersion()
}

// ClaimSchemaVersion is the schemaVersion value for CNAB claim documents.
func ClaimSchemaVersion() schema.Version {
	return cnabclaims.GetDefaultSchemaVersion()
}
