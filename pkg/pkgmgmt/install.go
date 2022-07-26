package pkgmgmt

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

type InstallOptions struct {
	PackageDownloadOptions

	Name          string
	URL           string
	FeedURL       string
	Version       string
	parsedURL     *url.URL
	parsedFeedURL *url.URL

	PackageType string
}

// GetParsedURL returns a copy of of the parsed URL that is safe to modify.
func (o *InstallOptions) GetParsedURL() url.URL {
	if o.parsedURL == nil {
		return url.URL{}
	}

	return *o.parsedURL
}

func (o *InstallOptions) GetParsedFeedURL() url.URL {
	if o.parsedFeedURL == nil {
		return o.defaultFeedURL()
	}

	return *o.parsedFeedURL
}

func (o *InstallOptions) defaultFeedURL() url.URL {
	mirror := o.GetMirror()
	mirror.Path = path.Join(mirror.Path, o.PackageType+"s", "atom.xml")
	return mirror
}

func (o *InstallOptions) Validate(args []string) error {
	if o.PackageType != "mixin" && o.PackageType != "plugin" {
		return fmt.Errorf("invalid package type %q. Please report this as a bug to Porter!", o.PackageType)
	}

	err := o.validateName(args)
	if err != nil {
		return err
	}

	err = o.PackageDownloadOptions.Validate()
	if err != nil {
		return err
	}

	err = o.validateFeedURL()
	if err != nil {
		return err
	}

	err = o.validateURL()
	if err != nil {
		return err
	}

	o.defaultVersion()

	return nil
}

func (o *InstallOptions) validateURL() error {
	if o.URL == "" {
		return nil
	}

	parsedURL, err := url.Parse(o.URL)
	if err != nil {
		return fmt.Errorf("invalid --url %s: %w", o.URL, err)
	}

	o.parsedURL = parsedURL
	return nil
}

func (o *InstallOptions) validateFeedURL() error {
	if o.URL == "" && o.FeedURL == "" {
		feedURL := o.defaultFeedURL()
		o.FeedURL = feedURL.String()
	}

	if o.FeedURL != "" {
		parsedFeedURL, err := url.Parse(o.FeedURL)
		if err != nil {
			return fmt.Errorf("invalid --feed-url %s: %w", o.FeedURL, err)
		}

		o.parsedFeedURL = parsedFeedURL
	}

	return nil
}

func (o *InstallOptions) defaultVersion() {
	if o.Version == "" {
		o.Version = "latest"
	}
}

// validateName grabs the name from the first positional argument.
func (o *InstallOptions) validateName(args []string) error {
	switch len(args) {
	case 0:
		return errors.New("no name was specified")
	case 1:
		o.Name = strings.ToLower(args[0])
		return nil
	default:
		return fmt.Errorf("only one positional argument may be specified, the name, but multiple were received: %s", args)

	}
}
