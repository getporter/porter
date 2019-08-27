package porter

import (
	"fmt"

	"github.com/deislabs/porter/pkg/printer"

	"github.com/deislabs/porter/pkg"
)

// VersionOptions represent generic options for use by Porter's list commands
type VersionOptions struct {
	printer.PrintOptions
}

var DefaultVersionFormat = printer.FormatPlaintext

func (o *VersionOptions) Validate() error {
	if o.RawFormat == "" {
		o.RawFormat = string(DefaultVersionFormat)
	}

	err := o.ParseFormat()
	if err != nil {
		return err
	}

	switch o.Format {
	case printer.FormatJson, printer.FormatPlaintext:
		return nil
	default:
		return fmt.Errorf("unsupported format, %s. Supported formats are: %s and %s", o.Format, printer.FormatJson, printer.FormatPlaintext)
	}
}

func (p *Porter) PrintVersion(opts VersionOptions) error {
	switch opts.Format {
	case printer.FormatJson:
		v := struct {
			Version string
			Commit  string
		}{
			pkg.Version,
			pkg.Commit,
		}
		return printer.PrintJson(p.Out, v)
	case printer.FormatPlaintext:
		_, err := fmt.Fprintf(p.Out, "porter %s (%s)\n", pkg.Version, pkg.Commit)
		return err
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}
