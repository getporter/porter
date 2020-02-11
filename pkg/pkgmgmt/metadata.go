package pkgmgmt

var _ PackageMetadata = Metadata{}

// Metadata about an installed package.
type Metadata struct {
	// Name of package.
	Name string `json:"name"`
	// VersionInfo for the package.
	VersionInfo
}

// GetName of the installed package.
func (m Metadata) GetName() string {
	return m.Name
}

// GetVersionInfo for the installed package.
func (m Metadata) GetVersionInfo() VersionInfo {
	return m.VersionInfo
}

// VersionInfo contains metadata from running the version command against the
// client executable.
type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Author  string `json:"author,omitempty"`
}

// PackageMetadata is a common interface for packages managed by Porter.
type PackageMetadata interface {
	// GetName of the installed package.
	GetName() string

	// GetVersionInfo for the installed package.
	GetVersionInfo() VersionInfo
}
