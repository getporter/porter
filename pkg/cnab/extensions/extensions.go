package extensions

import (
	"github.com/cnabio/cnab-go/bundle"
)

const (
	// PorterExtensionsPrefix is the prefix applied to any custom CNAB extensions developed by Porter.
	PorterExtensionsPrefix = "sh.porter."

	// OfficialExtensionsPrefix is the prefix applied to extensions defined in the CNAB spec.
	OfficialExtensionsPrefix = "io.cnab."
)

// SupportsExtension checks if the bundle supports the specified CNAB extension.
func SupportsExtension(b bundle.Bundle, key string) bool {
	for _, ext := range b.RequiredExtensions {
		if key == ext {
			return true
		}
	}
	return false
}
