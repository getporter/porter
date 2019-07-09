package porter

import (
	"fmt"
	"path/filepath"
	"time"

	dtprinter "github.com/carolynvs/datetime-printer"
	cnab "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/printer"
	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// ShowOptions represent options for showing a particular claim
type ShowOptions struct {
	RawFormat string
	Format    printer.Format
	Name      string
}

type ClaimListing struct {
	Name     string
	Created  time.Time
	Modified time.Time
	Action   string
	Status   string
	Outputs
}

// Validate prepares for a show bundle action and validates the options.
func (so *ShowOptions) Validate(args []string) error {
	if len(args) == 1 {
		so.Name = args[0]
	} else if len(args) > 1 {
		return errors.Errorf("only one positional argument may be specified, the claim name, but multiple were received: %s", args)
	}

	parsedFormat, err := printer.ParseFormat(so.RawFormat)
	if err != nil {
		return err
	}
	so.Format = parsedFormat

	return nil
}

// ShowBundle shows a bundle, or more properly a bundle claim, along with any
// associated outputs
func (p *Porter) ShowBundle(opts ShowOptions) error {
	cl, err := p.fetchClaim(opts.Name)
	if err != nil {
		return err
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

			// Get local output directory for this claim
			outputsDir, err := p.Config.GetOutputsDir()
			if err != nil {
				return errors.Wrap(err, "unable to get outputs directory")
			}
			claimOutputsDir := filepath.Join(outputsDir, cl.Name)

			var rows [][]string

			// Iterate through all Bundle Outputs and add to rows
			for _, o := range cl.Outputs {
				// TODO: test sensitive
				// TODO: test truncation
				value := o.Value
				// If output is sensitive, substitute local path
				if o.Sensitive {
					value = filepath.Join(claimOutputsDir, o.Name)
				}
				truncatedValue := truncateString(value, 60)
				rows = append(rows, []string{o.Name, o.Type, truncatedValue})
			}

			// Build and configure our tablewriter for the outputs
			table := tablewriter.NewWriter(p.Out)
			table.SetCenterSeparator("")
			table.SetColumnSeparator("")
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
			table.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: false, Top: true})
			table.SetAutoFormatHeaders(false)

			// Print the outputs table
			table.SetHeader([]string{"Name", "Type", "Value (Path if sensitive)"})
			for _, row := range rows {
				table.Append(row)
			}
			table.Render()
		}
		return nil

	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) fetchClaim(name string) (*ClaimListing, error) {
	cp := cnab.NewDuffle(p.Config)
	claimStore := cp.NewClaimStore()
	claim, err := claimStore.Read(name)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve claim %s", name)
	}

	var condensedClaims CondensedClaimList
	condensedClaim := CondensedClaim{
		Name:     claim.Name,
		Created:  claim.Created,
		Modified: claim.Modified,
		Action:   claim.Result.Action,
		Status:   claim.Result.Status,
	}
	condensedClaims = append(condensedClaims, condensedClaim)

	outputList, err := p.listBundleOutputs(name)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to list outputs for claim %s", name)
	}

	return &ClaimListing{
		Name:     claim.Name,
		Created:  claim.Created,
		Modified: claim.Modified,
		Action:   claim.Result.Action,
		Status:   claim.Result.Status,
		Outputs:  *outputList,
	}, nil

}
