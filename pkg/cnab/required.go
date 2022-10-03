package cnab

import "fmt"

// RequiredExtension represents a required extension that is known and supported by Porter
type RequiredExtension struct {
	Shorthand string
	Key       string
	Schema    string
	Reader    func(b ExtendedBundle) (interface{}, error)
}

// SupportedExtensions represent a listing of the current required extensions
// that Porter supports
var SupportedExtensions = []RequiredExtension{
	DependenciesV1Extension,
	DockerExtension,
	FileParameterExtension,
	ParameterSourcesExtension,
}

// ProcessedExtensions represents a map of the extension name to the
// processed extension configuration
type ProcessedExtensions map[string]interface{}

// ProcessRequiredExtensions checks all required extensions in the provided
// bundle and makes sure Porter supports them.
//
// If an unsupported required extension is found, an error is returned.
//
// For each supported required extension, the configuration for that extension
// is read and returned in the form of a map of the extension name to
// the extension configuration
func (b ExtendedBundle) ProcessRequiredExtensions() (ProcessedExtensions, error) {
	processed := ProcessedExtensions{}
	for _, reqExt := range b.RequiredExtensions {
		supportedExtension, err := GetSupportedExtension(reqExt)
		if err != nil {
			return processed, err
		}

		raw, err := supportedExtension.Reader(b)
		if err != nil {
			return processed, fmt.Errorf("unable to process extension: %s: %w", reqExt, err)
		}

		processed[supportedExtension.Key] = raw
	}

	return processed, nil
}

// GetSupportedExtension returns a supported extension according to the
// provided name, or an error
func GetSupportedExtension(e string) (*RequiredExtension, error) {
	for _, ext := range SupportedExtensions {
		// TODO(v1) we should only check for the key in v1.0.0
		// We are checking for both because of a bug in the cnab dependencies spec
		// https://github.com/cnabio/cnab-spec/issues/403
		if e == ext.Key || e == ext.Shorthand {
			return &ext, nil
		}
	}
	return nil, fmt.Errorf("unsupported required extension: %s", e)
}
