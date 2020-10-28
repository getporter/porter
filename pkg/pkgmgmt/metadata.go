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

// PackageListing represents discovery information for a package
type PackageListing struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	URL         string `json:"URL"`
}

// PackageList is a collection of PackageListings
type PackageList []PackageListing

// PackageList implements the sort.Interface for []PackageListing
// based on the Name field.
func (rml PackageList) Len() int {
	return len(rml)
}
func (rml PackageList) Swap(i, j int) {
	rml[i], rml[j] = rml[j], rml[i]
}
func (rml PackageList) Less(i, j int) bool {
	return rml[i].Name < rml[j].Name
}
