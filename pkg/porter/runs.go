package porter

import (
	"sort"
	"time"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/printer"
	dtprinter "github.com/carolynvs/datetime-printer"
)

// RunListOptions represent options for showing runs of an installation
type RunListOptions struct {
	sharedOptions
	printer.PrintOptions
}

// Validate prepares for the list installation runs action and validates the args/options.
func (so *RunListOptions) Validate(args []string, cxt *context.Context) error {
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

type DisplayRuns []DisplayRun

func (l DisplayRuns) Len() int {
	return len(l)
}

func (l DisplayRuns) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l DisplayRuns) Less(i, j int) bool {
	return l[i].Started.Before(l[j].Started)
}

func (p *Porter) ListInstallationRuns(opts RunListOptions) (DisplayRuns, error) {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return nil, err
	}

	var displayRuns DisplayRuns

	runs, runResults, err := p.Claims.ListRuns(opts.Namespace, opts.Name)
	if err != nil {
		return nil, err
	}

	for _, run := range runs {
		results := runResults[run.ID]

		displayRun := NewDisplayRun(run)

		if len(results) > 0 {
			displayRun.Status = results[len(results)-1].Status

			switch len(results) {
			case 2:
				displayRun.Started = results[0].Created
				displayRun.Stopped = results[1].Created
			case 1:
				displayRun.Started = results[0].Created
			default:
				displayRun.Stopped = results[len(results)-1].Created
			}
		}

		displayRuns = append(displayRuns, displayRun)
	}

	return displayRuns, nil
}

func (p *Porter) PrintInstallationRuns(opts RunListOptions) error {
	displayRuns, err := p.ListInstallationRuns(opts)
	if err != nil {
		return err
	}

	sort.Sort(sort.Reverse(displayRuns))

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, displayRuns)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, displayRuns)
	case printer.FormatPlaintext:
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		row :=
			func(v interface{}) []string {
				a, ok := v.(DisplayRun)
				if !ok {
					return nil
				}
				return []string{a.ClaimID, a.Action, tp.Format(a.Started), tp.Format(a.Stopped), a.Status}
			}
		return printer.PrintTable(p.Out, displayRuns, row, "Run ID", "Action", "Started", "Stopped", "Status")
	}

	return nil
}
