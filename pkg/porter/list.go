package porter

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
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
	Name              string
	Namespace         string
	Created           time.Time
	Modified          time.Time
	TrackedRepository string
	TrackedVersion    string
	TrackedDigest     string
	StatusLastAction  string
	StatusText        string
	StatusReference   string
	StatusVersion     string
	StatusDigest      string

	Parameters     DisplayValues
	Labels         []string
	ParameterSets  []string
	CredentialSets []string
}

func NewDisplayInstallation(installation claims.Installation, run *claims.Run) DisplayInstallation {
	di := DisplayInstallation{
		Name:              installation.Name,
		Namespace:         installation.Namespace,
		TrackedRepository: installation.BundleRepository,
		TrackedVersion:    installation.BundleVersion,
		TrackedDigest:     installation.BundleDigest,
		ParameterSets:     installation.ParameterSets,
		CredentialSets:    installation.CredentialSets,
		Created:           installation.Created,
		Modified:          installation.Modified,
		StatusReference:   installation.Status.BundleReference,
		StatusVersion:     installation.Status.BundleVersion,
		StatusDigest:      installation.Status.BundleDigest,
		StatusLastAction:  installation.Status.Action,
		StatusText:        installation.Status.ResultStatus,
	}

	labels := make([]string, 0, len(installation.Labels))
	for k, v := range installation.Labels {
		labels = append(labels, fmt.Sprintf("%s: %s", k, v))
	}
	sort.Strings(labels)
	di.Labels = labels

	// This is unset when we are just listing installations
	if run != nil {
		bun := cnab.ExtendedBundle{run.Bundle}
		di.Parameters = NewDisplayValuesFromParameters(bun, run.Parameters)
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
	return l[i].Modified.Before(l[j].Modified)
}

type DisplayRun struct {
	ClaimID    string                 `json:"claimID" yaml:"claimID"`
	Bundle     string                 `json:"bundle,omitempty" yaml:"bundle,omitempty"`
	Version    string                 `json:"version" yaml:"version"`
	Action     string                 `json:"action" yaml:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Timestamp  time.Time              `json:"timestamp" yaml:"timestamp"`
	Status     string                 `json:"status" yaml:"status"`
}

func NewDisplayRun(run claims.Run) DisplayRun {
	return DisplayRun{
		ClaimID:    run.ID,
		Action:     run.Action,
		Parameters: run.Parameters,
		Timestamp:  run.Created,
		Bundle:     run.BundleReference,
		Version:    run.Bundle.Version,
		// TODO(carolynvs): Add command to view all installation runs
		//Status: run.GetStatus(),
	}
}

// ListInstallations lists installed bundles.
func (p *Porter) ListInstallations(opts ListOptions) ([]claims.Installation, error) {
	installations, err := p.Claims.ListInstallations(opts.GetNamespace(), opts.Name, opts.ParseLabels())
	return installations, errors.Wrap(err, "could not list installations")
}

// PrintInstallations prints installed bundles.
func (p *Porter) PrintInstallations(opts ListOptions) error {
	installations, err := p.ListInstallations(opts)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, installations)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, installations)
	case printer.FormatTable:
		var displayInstallations DisplayInstallations
		for _, installation := range installations {
			displayInstallations = append(displayInstallations, NewDisplayInstallation(installation, nil))
		}
		sort.Sort(sort.Reverse(displayInstallations))

		// have every row use the same "now" starting ... NOW!
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		row :=
			func(v interface{}) []interface{} {
				cl, ok := v.(DisplayInstallation)
				if !ok {
					return nil
				}
				return []interface{}{cl.Namespace, cl.Name, tp.Format(cl.Created), tp.Format(cl.Modified), cl.StatusLastAction, cl.StatusText}
			}
		return printer.PrintTable(p.Out, displayInstallations, row,
			"NAMESPACE", "NAME", "CREATED", "MODIFIED", "LAST ACTION", "LAST STATUS")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
