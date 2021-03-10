package porter

import (
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/printer"
	dtprinter "github.com/carolynvs/datetime-printer"
)

var (
	ShowAllowedFormats = []printer.Format{printer.FormatTable, printer.FormatYaml, printer.FormatJson}
	ShowDefaultFormat  = printer.FormatTable
)

// ShowOptions represent options for showing a particular installation
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

	return so.PrintOptions.Validate(ShowDefaultFormat, ShowAllowedFormats)
}

// GetInstallation retrieves information about an installation.
func (p *Porter) GetInstallation(opts ShowOptions) (DisplayInstallation, error) {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return DisplayInstallation{}, err
	}

	installation, err := p.Claims.ReadInstallation(opts.Name)
	if err != nil {
		return DisplayInstallation{}, err
	}

	outputs, err := p.Claims.ReadLastOutputs(opts.Name)
	if err != nil {
		return DisplayInstallation{}, err
	}

	displayInstallation, err := NewDisplayInstallation(installation)
	if err != nil {
		// There isn't an installation to display
		return DisplayInstallation{}, err
	}

	c, err := installation.GetLastClaim()
	if err != nil {
		return DisplayInstallation{}, err
	}
	displayInstallation.Outputs = NewDisplayOutputs(c.Bundle, outputs, opts.Format)

	return displayInstallation, nil
}

// ShowInstallation shows a bundle installation, along with any
// associated outputs
func (p *Porter) ShowInstallation(opts ShowOptions) error {
	displayInstallation, err := p.GetInstallation(opts)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, displayInstallation)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, displayInstallation)
	case printer.FormatTable:
		// Set up human friendly time formatter
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		// Print installation details
		fmt.Fprintf(p.Out, "Name: %s\n", displayInstallation.Name)
		fmt.Fprintf(p.Out, "Created: %s\n", tp.Format(displayInstallation.Created))
		fmt.Fprintf(p.Out, "Modified: %s\n", tp.Format(displayInstallation.Modified))

		// Print outputs, if any
		if len(displayInstallation.Outputs) > 0 {
			fmt.Fprintln(p.Out)
			fmt.Fprintln(p.Out, "Outputs:")

			err = p.printOutputsTable(displayInstallation.Outputs)
			if err != nil {
				return err
			}
		}

		fmt.Fprintln(p.Out)
		fmt.Fprintln(p.Out, "History:")
		historyRow :=
			func(v interface{}) []string {
				a, ok := v.(InstallationAction)
				if !ok {
					return nil
				}
				return []string{a.ClaimID, a.Action, tp.Format(a.Timestamp), a.Status, a.HasLogs}
			}
		return printer.PrintTableSection(p.Out, displayInstallation.History, historyRow, "Run ID", "Action", "Timestamp", "Status", "Has Logs")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
