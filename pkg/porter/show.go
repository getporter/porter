package porter

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	dtprinter "github.com/carolynvs/datetime-printer"
	tablewriter "github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/config"
	context "github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/printer"
)

// ShowOptions represent options for showing a particular claim
type ShowOptions struct {
	sharedOptions
	RawFormat string
	Format    printer.Format
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
	err := so.sharedOptions.validateClaimName(args)
	if err != nil {
		return err
	}

	// If claim name not supplied, try to determine from manifest in current working directory
	if so.sharedOptions.Name == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "could not get current working directory")
		}

		manifestExists, err := cxt.FileSystem.Exists(filepath.Join(pwd, config.Name))
		if err != nil {
			return errors.Wrap(err, "could not check if porter manifest exists in current directory")
		}

		if manifestExists {
			so.sharedOptions.File = config.Name
		} else {
			return errors.New("claim name must be provided")
		}
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
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}
	name := opts.sharedOptions.Name

	claim, err := p.CNAB.FetchClaim(name)
	if err != nil {
		return err
	}

	outputs, err := p.listBundleOutputs(name)
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

			// Get local output directory for this claim
			outputsDir, err := p.Config.GetOutputsDir()
			if err != nil {
				return errors.Wrap(err, "unable to get outputs directory")
			}
			claimOutputsDir := filepath.Join(outputsDir, cl.Name)

			var rows [][]string

			// Iterate through all Bundle Outputs and add to rows
			for _, o := range cl.Outputs {
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
