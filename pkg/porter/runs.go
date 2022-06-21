package porter

import (
	"context"
	"sort"
	"time"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
	dtprinter "github.com/carolynvs/datetime-printer"
)

// RunListOptions represent options for showing runs of an installation
type RunListOptions struct {
	sharedOptions
	printer.PrintOptions
}

// Validate prepares for the list installation runs action and validates the args/options.
func (so *RunListOptions) Validate(args []string, cxt *portercontext.Context) error {
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

func (p *Porter) ListInstallationRuns(ctx context.Context, opts RunListOptions) (DisplayRuns, error) {
	err := p.applyDefaultOptions(ctx, &opts.sharedOptions)
	if err != nil {
		return nil, err
	}

	var displayRuns DisplayRuns

	runs, runResults, err := p.Installations.ListRuns(ctx, opts.Namespace, opts.Name)
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
				displayRun.Stopped = &results[1].Created
			case 1:
				displayRun.Started = results[0].Created
			default:
				displayRun.Stopped = &results[len(results)-1].Created
			}
		}

		displayRuns = append(displayRuns, displayRun)
	}

	return displayRuns, nil
}

func (p *Porter) PrintInstallationRuns(ctx context.Context, opts RunListOptions) error {
	displayRuns, err := p.ListInstallationRuns(ctx, opts)
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

				stopped := ""
				if a.Stopped != nil {
					stopped = tp.Format(*a.Stopped)
				}

				return []string{a.ID, a.Action, tp.Format(a.Started), stopped, a.Status}
			}
		return printer.PrintTable(p.Out, displayRuns, row, "Run ID", "Action", "Started", "Stopped", "Status")
	}

	return nil
}
