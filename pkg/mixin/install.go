package mixin

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

type InstallOptions struct {
	Name      string
	URL       string
	Version   string
	parsedURL *url.URL
}

// CopyParsedURL returns a copy of of the parsed URL that is safe to modify.
func (o *InstallOptions) GetParsedURL() url.URL {
	return *o.parsedURL
}

func (o *InstallOptions) Validate(args []string) error {
	err := o.validateMixinName(args)
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
	parsedURL, err := url.Parse(o.URL)
	if err != nil {
		return errors.Wrapf(err, "invalid --url %s", o.URL)
	}

	o.parsedURL = parsedURL
	return nil
}

func (o *InstallOptions) defaultVersion() {
	if o.Version == "" {
		o.Version = "latest"
	}
}

// validateClaimName grabs the claim name from the first positional argument.
func (o *InstallOptions) validateMixinName(args []string) error {
	switch len(args) {
	case 0:
		// later on this may be okay (porter mixin install --feed deislabs) but not now
		return errors.Errorf("no mixin name was specified")
	case 1:
		o.Name = strings.ToLower(args[0])
		return nil
	default:
		return errors.Errorf("only one positional argument may be specified, the mixin name, but multiple were received: %s", args)

	}
}
