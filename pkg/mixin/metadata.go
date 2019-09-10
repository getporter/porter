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

// Information from running the version command against the mixin.
type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Author  string `json:"author,omitempty"`
}
