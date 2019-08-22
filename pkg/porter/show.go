package porter

import (
	"fmt"
	"time"

	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/printer"
)

// ShowOptions represent options for showing a particular claim
type ShowOptions struct {
	sharedOptions
	printer.PrintOptions
}

// Validate prepares for a show bundle action and validates the args/options.
func (so *ShowOptions) Validate(args []string, cxt *context.Context) error {
	// Ensure only one argument exists (instance name) if args length non-zero
	err := so.sharedOptions.validateInstanceName(args)
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
func (p *Porter) ShowInstances(opts ShowOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}
	name := opts.sharedOptions.Name

	claim, err := p.CNAB.FetchClaim(name)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, claim)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, claim)
	case printer.FormatTable:
		// Set up human friendly time formatter
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		// Print claim details
		fmt.Fprintf(p.Out, "Name: %s\n", claim.Name)
		fmt.Fprintf(p.Out, "Created: %s\n", tp.Format(claim.Created))
		fmt.Fprintf(p.Out, "Modified: %s\n", tp.Format(claim.Modified))
		fmt.Fprintf(p.Out, "Last Action: %s\n", claim.Result.Action)
		fmt.Fprintf(p.Out, "Last Status: %s\n", claim.Result.Status)

		// Print outputs, if any
		if len(claim.Outputs) > 0 {
			fmt.Fprintln(p.Out)
			fmt.Fprint(p.Out, "Outputs:\n")

			// Fetch full Output objects with various metadata fields used for printing
			outputObjects, err := p.fetchBundleOutputs(name)
			if err != nil {
				return err
			}

			return p.printOutputsTable(outputObjects, claim.Name)
		}
		return nil
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
