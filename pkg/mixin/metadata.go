package mixin

// Metadata about an installed mixin.
type Metadata struct {
	// Mixin Name
	Name string `json:"name"`
	// Mixin Directory
	Dir string `json:"dir,omitempty"`
	// Path to the client executable
	ClientPath string `json:"clientPath,omitempty"`
	// Metadata about the mixin version returned from calling version on the mixin
	VersionInfo
}

// VersionInfo represents information from running the version command against the mixin.
type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Author  string `json:"author,omitempty"`
}

// RemoteMixinListing represents an informational listing for a remote mixin
type RemoteMixinListing struct {
	Name        string `json:"name"`
	Author      string `json:"author"`
	Description string `json:"description"`
	URL         string `json:"URL"`
}

// RemoteMixinList is a collection of RemoteMixinListings
type RemoteMixinList []RemoteMixinListing

// RemoteMixinList implements the sort.Interface for []RemoteMixinListing
// based on the Name field.
func (rml RemoteMixinList) Len() int {
	return len(rml)
}
func (rml RemoteMixinList) Swap(i, j int) {
	rml[i], rml[j] = rml[j], rml[i]
}
func (rml RemoteMixinList) Less(i, j int) bool {
	return rml[i].Name < rml[j].Name
}
