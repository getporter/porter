package porter

import (
	"fmt"
	"time"

	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/pkg/errors"

	context "github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/printer"
)

// ShowOptions represent options for showing a particular claim
type ShowOptions struct {
	sharedOptions
	printer.PrintOptions
}

type ClaimListing struct {
	Name     string
	Created  time.Time
	Modified time.Time
	Action   string
	Status   string
	Outputs
}

// Validate prepares for a show bundle action and validates the args/options.
func (so *ShowOptions) Validate(args []string, cxt *context.Context) error {
	// Ensure only one argument exists (claim name) if args length non-zero
	err := so.sharedOptions.validateClaimName(args)
	if err != nil {
		return err
	}

	// If args length zero, attempt to derive claim name from context
	err = so.sharedOptions.defaultBundleFiles(cxt)
	if err != nil {
		return errors.Wrap(err, "claim name must be provided")
	}

	return so.ParseFormat()
}

// ShowBundle shows a bundle, or more properly a bundle claim, along with any
// associated outputs
func (p *Porter) ShowBundle(opts ShowOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}
	name := opts.sharedOptions.Name

	claim, err := p.CNAB.FetchClaim(name)
	if err != nil {
		return err
	}

	outputs, err := p.fetchBundleOutputs(name)
	if err != nil {
		return err
	}

	cl := ClaimListing{
		Name:     claim.Name,
		Created:  claim.Created,
		Modified: claim.Modified,
		Action:   claim.Result.Action,
		Status:   claim.Result.Status,
		Outputs:  *outputs,
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, cl)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, cl)
	case printer.FormatTable:
		// Set up human friendly time formatter
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		// Print claim details
		fmt.Fprintf(p.Out, "Name: %s\n", cl.Name)
		fmt.Fprintf(p.Out, "Created: %s\n", tp.Format(cl.Created))
		fmt.Fprintf(p.Out, "Modified: %s\n", tp.Format(cl.Modified))
		fmt.Fprintf(p.Out, "Last Action: %s\n", cl.Action)
		fmt.Fprintf(p.Out, "Last Status: %s\n", cl.Status)

		// Print outputs, if any
		if cl.Outputs.Len() > 0 {
			fmt.Fprintln(p.Out)
			fmt.Fprint(p.Out, "Outputs:\n")

			return p.printOutputsTable(&cl.Outputs, cl.Name)
		}
		return nil
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
