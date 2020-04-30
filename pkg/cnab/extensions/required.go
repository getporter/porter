package extensions

import (
	"github.com/cnabio/cnab-go/bundle"
	"github.com/pkg/errors"
)

// RequiredExtension represents a required extension that is known and supported by Porter
type RequiredExtension struct {
	Shorthand string
	Key       string
	Schema    string
	Reader    func(b bundle.Bundle) (interface{}, error)
}

// SupportedExtensions represent a listing of the current required extensions
// that Porter supports
var SupportedExtensions = []RequiredExtension{
	DependenciesExtension,
	DockerExtension,
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
func ProcessRequiredExtensions(b bundle.Bundle) (ProcessedExtensions, error) {
	processed := ProcessedExtensions{}
	for _, reqExt := range b.RequiredExtensions {
		supportedExtension, err := GetSupportedExtension(reqExt)
		if err != nil {
			return processed, err
		}

		raw, err := supportedExtension.Reader(b)
		if err != nil {
			return processed, errors.Wrapf(err, "unable to process extension: %s", reqExt)
		}

		processed[supportedExtension.Key] = raw
	}

	return processed, nil
}

// GetSupportedExtension returns a supported extension according to the
// provided name, or an error
func GetSupportedExtension(e string) (*RequiredExtension, error) {
	for _, ext := range SupportedExtensions {
		if e == ext.Key || e == ext.Shorthand {
			return &ext, nil
		}
	}
	return nil, errors.Errorf("unsupported required extension: %s", e)
}
