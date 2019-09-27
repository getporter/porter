package version

import (
	"fmt"
	"os"

	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/deislabs/porter/pkg/printer"
)

// VersionOptions represent generic options for use by version commands
type Options struct {
	printer.PrintOptions
}

var DefaultVersionFormat = printer.FormatPlaintext
var GetExecutable = os.Executable

//
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
// Suitable for any mixin to use to implement its version command.
func PrintVersion(cxt *context.Context, opts Options, metadata mixin.Metadata) error {
	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(cxt.Out, metadata)
	case printer.FormatPlaintext:
		authorship := ""
		if metadata.VersionInfo.Author != "" {
			authorship = " by " + metadata.VersionInfo.Author
		}
		_, err := fmt.Fprintf(cxt.Out, "%s %s (%s)%s\n", metadata.Name, metadata.VersionInfo.Version, metadata.VersionInfo.Commit, authorship)
		return err
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}
