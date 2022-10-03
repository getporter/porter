package pkgmgmt

import (
	"fmt"
	"net/url"
)

// DefaultPackageMirror is the default location from which to download Porter assets, such as binaries, atom feeds and package indexes.
const DefaultPackageMirror = "https://cdn.porter.sh"

// GetDefaultPackageMirrorURL returns DefaultPackageMirror parsed as a url.URL
func GetDefaultPackageMirrorURL() url.URL {
	defaultMirror, _ := url.Parse(DefaultPackageMirror)
	return *defaultMirror
}

// PackageDownloadOptions are options for downloading Porter packages, such as mixins and plugins.
type PackageDownloadOptions struct {
	Mirror       string
	parsedMirror *url.URL
}

func (o *PackageDownloadOptions) Validate() error {
	if o.Mirror == "" {
		o.Mirror = DefaultPackageMirror
	}

	mirrorURL, err := url.Parse(o.Mirror)
	if err != nil {
		return fmt.Errorf("invalid --mirror %s: %w", o.Mirror, err)
	}
	o.parsedMirror = mirrorURL

	return nil
}

// GetMirror returns a copy of of the parsed Mirror that is safe to modify.
func (o *PackageDownloadOptions) GetMirror() url.URL {
	if o.parsedMirror == nil {
		return GetDefaultPackageMirrorURL()
	}

	return *o.parsedMirror
}
