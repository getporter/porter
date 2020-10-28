package porter

import (
	"fmt"
	"time"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/printer"
	dtprinter "github.com/carolynvs/datetime-printer"
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

	return so.ParseFormat()
}

// ShowInstallation shows a bundle installation, along with any
// associated outputs
func (p *Porter) ShowInstallation(opts ShowOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	installation, err := p.Claims.ReadInstallation(opts.Name)
	if err != nil {
		return err
	}

	outputs, err := p.Claims.ReadLastOutputs(opts.Name)
	if err != nil {
		return err
	}

	displayInstallation, err := NewDisplayInstallation(installation)
	if err != nil {
		// There isn't an installation to display
		return err
	}

	displayInstallation.Outputs = NewDisplayOutputs(outputs, opts.Format)

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, installation)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, installation)
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
				return []string{a.Action, tp.Format(a.Timestamp), a.Status}
			}
		return printer.PrintTableSection(p.Out, displayInstallation.History, historyRow, "Action", "Timestamp", "Status")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
