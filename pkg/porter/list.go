package porter

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/tracing"
	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/pkg/errors"
)

// ListOptions represent generic options for use by Porter's list commands
type ListOptions struct {
	printer.PrintOptions
	AllNamespaces bool
	Namespace     string
	Name          string
	Labels        []string
}

func (o *ListOptions) Validate() error {
	return o.ParseFormat()
}

func (o ListOptions) GetNamespace() string {
	if o.AllNamespaces {
		return "*"
	}
	return o.Namespace
}

func (o ListOptions) ParseLabels() map[string]string {
	return parseLabels(o.Labels)
}

func parseLabels(raw []string) map[string]string {
	if len(raw) == 0 {
		return nil
	}

	labelMap := make(map[string]string, len(raw))
	for _, label := range raw {
		parts := strings.SplitN(label, "=", 2)
		k := parts[0]
		v := ""
		if len(parts) > 1 {
			v = parts[1]
		}
		labelMap[k] = v
	}
	return labelMap
}

// DisplayInstallation holds a subset of pertinent values to be listed from installation data
// originating from its claims, results and outputs records
type DisplayInstallation struct {
	// SchemaType helps when we export the definition so editors can detect the type of document, it's not used by porter.
	SchemaType                  string `json:"schemaType" yaml:"schemaType"`
	claims.Installation         `yaml:",inline"`
	DisplayInstallationMetadata `json:"_calculated" yaml:"_calculated"`
}

type DisplayInstallationMetadata struct {
	ResolvedParameters DisplayValues `json:"resolvedParameters", yaml:"resolvedParameters"`
}

func NewDisplayInstallation(installation claims.Installation, run *claims.Run) DisplayInstallation {
	di := DisplayInstallation{
		SchemaType:   "Installation",
		Installation: installation,
	}

	// This is unset when we are just listing installations
	if run != nil {
		bun := cnab.ExtendedBundle{run.Bundle}
		di.ResolvedParameters = NewDisplayValuesFromParameters(bun, run.Parameters)
	}

	return di
}

// TODO(carolynvs): be consistent with sorting results from list, either keep the default sort by name
// or update the other types to also sort by modified
type DisplayInstallations []DisplayInstallation

func (l DisplayInstallations) Len() int {
	return len(l)
}

func (l DisplayInstallations) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l DisplayInstallations) Less(i, j int) bool {
	return l[i].Status.Modified.Before(l[j].Status.Modified)
}

type DisplayRun struct {
	ClaimID    string                 `json:"claimID" yaml:"claimID"`
	Bundle     string                 `json:"bundle,omitempty" yaml:"bundle,omitempty"`
	Version    string                 `json:"version" yaml:"version"`
	Action     string                 `json:"action" yaml:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Started    time.Time              `json:"started" yaml:"started"`
	Stopped    time.Time              `json:"stopped" yaml:"stopped"`
	Status     string                 `json:"status" yaml:"status"`
}

func NewDisplayRun(run claims.Run) DisplayRun {
	return DisplayRun{
		ClaimID:    run.ID,
		Action:     run.Action,
		Parameters: run.Parameters,
		Started:    run.Created,
		Bundle:     run.BundleReference,
		Version:    run.Bundle.Version,
	}
}

// ListInstallations lists installed bundles.
func (p *Porter) ListInstallations(ctx context.Context, opts ListOptions) ([]claims.Installation, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	installations, err := p.Claims.ListInstallations(opts.GetNamespace(), opts.Name, opts.ParseLabels())
	return installations, errors.Wrap(err, "could not list installations")
}

// PrintInstallations prints installed bundles.
func (p *Porter) PrintInstallations(ctx context.Context, opts ListOptions) error {
	installations, err := p.ListInstallations(ctx, opts)
	if err != nil {
		return err
	}

	var displayInstallations DisplayInstallations
	for _, installation := range installations {
		displayInstallations = append(displayInstallations, NewDisplayInstallation(installation, nil))
	}
	sort.Sort(sort.Reverse(displayInstallations))

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, displayInstallations)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, displayInstallations)
	case printer.FormatPlaintext:
		// have every row use the same "now" starting ... NOW!
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		row :=
			func(v interface{}) []string {
				cl, ok := v.(DisplayInstallation)
				if !ok {
					return nil
				}
				return []string{cl.Namespace, cl.Name, tp.Format(cl.Status.Created), tp.Format(cl.Status.Modified), cl.Status.Action, cl.Status.ResultStatus}
			}
		return printer.PrintTable(p.Out, displayInstallations, row,
			"NAMESPACE", "NAME", "CREATED", "MODIFIED", "LAST ACTION", "LAST STATUS")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
