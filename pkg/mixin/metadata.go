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
	SourceURL   string `json:"sourceURL"`
	FeedURL     string `json:"feedURL"`
}

// RemoteMixinList is a collection of RemoteMixinListings
type RemoteMixinList []RemoteMixinListing

func (rms RemoteMixinList) Len() int {
	return len(rms)
}
func (rms RemoteMixinList) Swap(i, j int) {
	rms[i], rms[j] = rms[j], rms[i]
}
func (rms RemoteMixinList) Less(i, j int) bool {
	return rms[i].Name < rms[j].Name
}
