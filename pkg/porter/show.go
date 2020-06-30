package porter

import (
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/printer"
	dtprinter "github.com/carolynvs/datetime-printer"
)

// ShowOptions represent options for showing a particular claim
type ShowOptions struct {
	sharedOptions
	printer.PrintOptions
}

// Validate prepares for a show bundle action and validates the args/options.
func (so *ShowOptions) Validate(args []string, cxt *context.Context) error {
	// Ensure only one argument exists (installation name) if args length non-zero
	err := so.sharedOptions.validateInstallationName(args)
	if err != nil {
		return err
	}

	err = so.sharedOptions.defaultBundleFiles(cxt)
	if err != nil {
		return err
	}

	return so.ParseFormat()
}

// ShowInstances shows a bundle, or more properly a bundle claim, along with any
// associated outputs
func (p *Porter) ShowInstallations(opts ShowOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	c, err := p.Claims.Read(opts.sharedOptions.Name)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, c)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, c)
	case printer.FormatTable:
		// Set up human friendly time formatter
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		// Print claim details
		fmt.Fprintf(p.Out, "Name: %s\n", c.Installation)
		fmt.Fprintf(p.Out, "Created: %s\n", tp.Format(c.Created))
		fmt.Fprintf(p.Out, "Modified: %s\n", tp.Format(c.Modified))
		fmt.Fprintf(p.Out, "Last Action: %s\n", c.Result.Action)
		fmt.Fprintf(p.Out, "Last Status: %s\n", c.Result.Status)

		// Print outputs, if any
		if len(c.Outputs) > 0 {
			fmt.Fprintln(p.Out)
			fmt.Fprint(p.Out, "Outputs:\n")

			outputs := p.ListBundleOutputs(c, opts.Format)
			return p.printOutputsTable(outputs)
		}
		return nil
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
