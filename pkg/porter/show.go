package porter

import (
	"fmt"
	"sort"
	"time"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/printer"
	dtprinter "github.com/carolynvs/datetime-printer"
)

var (
	ShowAllowedFormats = []printer.Format{printer.FormatPlaintext, printer.FormatYaml, printer.FormatJson}
	ShowDefaultFormat  = printer.FormatPlaintext
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

	requireBundle := so.Name == ""
	err = so.sharedOptions.defaultBundleFiles(cxt, requireBundle)
	if err != nil {
		return err
	}

	return so.PrintOptions.Validate(ShowDefaultFormat, ShowAllowedFormats)
}

// GetInstallation retrieves information about an installation, including its most recent run.
func (p *Porter) GetInstallation(opts ShowOptions) (claims.Installation, *claims.Run, error) {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return claims.Installation{}, nil, err
	}

	installation, err := p.Claims.GetInstallation(opts.Namespace, opts.Name)
	if err != nil {
		return claims.Installation{}, nil, err
	}

	if installation.Status.RunID != "" {
		run, err := p.Claims.GetRun(installation.Status.RunID)
		if err != nil {
			return claims.Installation{}, nil, err
		}
		return installation, &run, nil
	}

	return installation, nil, nil
}

// ShowInstallation shows a bundle installation, along with any
// associated outputs
func (p *Porter) ShowInstallation(opts ShowOptions) error {
	installation, run, err := p.GetInstallation(opts)
	if err != nil {
		return err
	}

	displayInstallation := NewDisplayInstallation(installation, run)

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, displayInstallation)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, displayInstallation)
	case printer.FormatPlaintext:
		// Set up human friendly time formatter
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		// Print installation details
		fmt.Fprintf(p.Out, "Name: %s\n", displayInstallation.Name)
		fmt.Fprintf(p.Out, "Namespace: %s\n", displayInstallation.Namespace)
		fmt.Fprintf(p.Out, "Created: %s\n", tp.Format(displayInstallation.Status.Created))
		fmt.Fprintf(p.Out, "Modified: %s\n", tp.Format(displayInstallation.Status.Modified))

		if displayInstallation.Bundle.Repository != "" {
			fmt.Fprintln(p.Out)
			fmt.Fprintln(p.Out, "Bundle:")
			fmt.Fprintf(p.Out, "  Repository: %s\n", displayInstallation.Bundle.Repository)
			if displayInstallation.Bundle.Version != "" {
				fmt.Fprintf(p.Out, "  Version: %s\n", displayInstallation.Bundle.Version)
			}
			if displayInstallation.Bundle.Digest != "" {
				fmt.Fprintf(p.Out, "  Digest: %s\n", displayInstallation.Bundle.Digest)
			}
		}

		// Print labels, if any
		if len(displayInstallation.Labels) > 0 {
			fmt.Fprintln(p.Out)
			fmt.Fprintln(p.Out, "Labels:")

			// Print labels in alphabetical order
			labels := make([]string, 0, len(installation.Labels))
			for k, v := range installation.Labels {
				labels = append(labels, fmt.Sprintf("%s: %s", k, v))
			}
			sort.Strings(labels)

			for _, label := range labels {
				fmt.Fprintf(p.Out, "  %s\n", label)
			}
		}

		// Print parameters, if any
		if len(displayInstallation.Parameters) > 0 {
			fmt.Fprintln(p.Out)
			fmt.Fprintln(p.Out, "Parameters:")

			err = p.printDisplayValuesTable(displayInstallation.ResolvedParameters)
			if err != nil {
				return err
			}
		}

		// Print parameter sets, if any
		if len(displayInstallation.ParameterSets) > 0 {
			fmt.Fprintln(p.Out)
			fmt.Fprintln(p.Out, "Parameter Sets:")
			for _, ps := range displayInstallation.ParameterSets {
				fmt.Fprintf(p.Out, "  - %s\n", ps)
			}
		}

		// Print credential sets, if any
		if len(displayInstallation.CredentialSets) > 0 {
			fmt.Fprintln(p.Out)
			fmt.Fprintln(p.Out, "Credential Sets:")
			for _, cs := range displayInstallation.CredentialSets {
				fmt.Fprintf(p.Out, "  - %s\n", cs)
			}
		}

		// Print the status (it may not be present if it's newly created using apply)
		if installation.Status != (claims.InstallationStatus{}) {
			fmt.Fprintln(p.Out)
			fmt.Fprintln(p.Out, "Status:")
			fmt.Fprintf(p.Out, "  Reference: %s\n", displayInstallation.Status.BundleReference)
			fmt.Fprintf(p.Out, "  Version: %s\n", displayInstallation.Status.BundleVersion)
			fmt.Fprintf(p.Out, "  Last Action: %s\n", displayInstallation.Status.Action)
			fmt.Fprintf(p.Out, "  Status: %s\n", displayInstallation.Status.ResultStatus)
			fmt.Fprintf(p.Out, "  Digest: %s\n", displayInstallation.Status.BundleDigest)
		}

		return nil
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
