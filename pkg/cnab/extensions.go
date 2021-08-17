package cnab

const (
	// PorterExtension is the key for all Porter configuration stored the the custom section of bundles.
	PorterExtension = "sh.porter"

	// PorterExtensionsPrefix is the prefix applied to any custom CNAB extensions developed by Porter.
	PorterExtensionsPrefix = PorterExtension + "."

	// OfficialExtensionsPrefix is the prefix applied to extensions defined in the CNAB spec.
	OfficialExtensionsPrefix = "io.cnab."

	// PorterInternal is the identifier that we put in the $comment of fields in bundle.json
	// to indicate that it's just for Porter and shouldn't be visible to the end users.
	PorterInternal = "porter-internal"
)

// SupportsExtension checks if the bundle supports the specified CNAB extension.
func (b ExtendedBundle) SupportsExtension(key string) bool {
	for _, ext := range b.RequiredExtensions {
		if key == ext {
			return true
		}
	}
	return false
}
