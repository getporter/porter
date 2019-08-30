package definition

import (
	"github.com/qri-io/jsonschema"
)

// ContentEncoding represents a "custom" Schema property
type ContentEncoding string

// NewContentEncoding allocates a new ContentEncoding validator
func NewContentEncoding() jsonschema.Validator {
	return new(ContentEncoding)
}

// Validate implements the Validator interface for ContentEncoding
// Currently, this is a no-op and is only used to register with the jsonschema library
// that 'contentEncoding' is a valid property (as of writing, it isn't one of the defaults)
func (c ContentEncoding) Validate(propPath string, data interface{}, errs *[]jsonschema.ValError) {}

// NewRootSchema returns a jsonschema.RootSchema with any needed custom
// jsonschema.Validators pre-registered
func NewRootSchema() *jsonschema.RootSchema {
	// Register custom validators here
	// Note: as of writing, jsonschema doesn't have a stock validator for intances of type `contentEncoding`
	// There may be others missing in the library that exist in http://json-schema.org/draft-07/schema#
	// and thus, we'd need to create/register them here (if not included upstream)
	jsonschema.RegisterValidator("contentEncoding", NewContentEncoding)
	return new(jsonschema.RootSchema)
}
