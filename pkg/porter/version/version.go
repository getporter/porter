package version

import (
	"fmt"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
)

// VersionOptions represent generic options for use by version commands
type Options struct {
	printer.PrintOptions
}

var DefaultVersionFormat = printer.FormatPlaintext

func (o *Options) Validate() error {
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

// PrintVersion prints the version based on the version flags using the binary's metadata.
// Suitable for any mixin or plugin to use to implement its version command.
func PrintVersion(cxt *portercontext.Context, opts Options, metadata pkgmgmt.PackageMetadata) error {
	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(cxt.Out, metadata)
	case printer.FormatPlaintext:
		vi := metadata.GetVersionInfo()
		authorship := ""
		if vi.Author != "" {
			authorship = " by " + vi.Author
		}
		_, err := fmt.Fprintf(cxt.Out, "%s %s (%s)%s\n", metadata.GetName(), vi.Version, vi.Commit, authorship)
		return err
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}
