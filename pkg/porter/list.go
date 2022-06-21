package porter

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/tracing"
	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/cnabio/cnab-go/schema"
	"github.com/pkg/errors"
)

// ListOptions represent generic options for use by Porter's list commands
type ListOptions struct {
	printer.PrintOptions
	AllNamespaces bool
	Namespace     string
	Name          string
	Labels        []string
	Skip          int64
	Limit         int64
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
// originating from its runs, results and outputs records
type DisplayInstallation struct {
	// SchemaType helps when we export the definition so editors can detect the type of document, it's not used by porter.
	SchemaType string `json:"schemaType" yaml:"schemaType" toml:"schemaType"`

	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`

	ID string `json:"id" yaml:"id" toml:"id"`
	// Name of the installation. Immutable.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Namespace in which the installation is defined.
	Namespace string `json:"namespace" yaml:"namespace" toml:"namespace"`

	// Uninstalled specifies if the installation isn't used anymore and should be uninstalled.
	Uninstalled bool `json:"uninstalled,omitempty" yaml:"uninstalled,omitempty" toml:"uninstalled,omitempty"`

	// Bundle specifies the bundle reference to use with the installation.
	Bundle storage.OCIReferenceParts `json:"bundle" yaml:"bundle" toml:"bundle"`

	// Custom extension data applicable to a given runtime.
	// TODO(carolynvs): remove and populate in ToCNAB when we firm up the spec
	Custom interface{} `json:"custom,omitempty" yaml:"custom,omitempty" toml:"custom,omitempty"`

	// Labels applied to the installation.
	Labels map[string]string `json:"labels,omitempty" yaml:"labels,omitempty" toml:"labels,omitempty"`

	// CredentialSets that should be included when the bundle is reconciled.
	CredentialSets []string `json:"credentialSets,omitempty" yaml:"credentialSets,omitempty" toml:"credentialSets,omitempty"`

	// Parameters specified by the user through overrides.
	// Does not include defaults, or values resolved from parameter sources.
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty" toml:"parameters,omitempty"`

	// ParameterSets that should be included when the bundle is reconciled.
	ParameterSets []string `json:"parameterSets,omitempty" yaml:"parameterSets,omitempty" toml:"parameterSets,omitempty"`

	// Status of the installation.
	Status                      storage.InstallationStatus `json:"status,omitempty" yaml:"status,omitempty" toml:"status,omitempty"`
	DisplayInstallationMetadata `json:"_calculated" yaml:"_calculated"`
}

type DisplayInstallationMetadata struct {
	ResolvedParameters DisplayValues `json:"resolvedParameters" yaml:"resolvedParameters"`
}

func NewDisplayInstallation(installation storage.Installation) DisplayInstallation {

	di := DisplayInstallation{
		SchemaType:     "Installation",
		SchemaVersion:  installation.SchemaVersion,
		ID:             installation.ID,
		Name:           installation.Name,
		Namespace:      installation.Namespace,
		Uninstalled:    installation.Uninstalled,
		Bundle:         installation.Bundle,
		Custom:         installation.Custom,
		Labels:         installation.Labels,
		CredentialSets: installation.CredentialSets,
		ParameterSets:  installation.ParameterSets,
		Status:         installation.Status,
	}

	return di
}

// ConvertToInstallationClaim transforms the data from DisplayInstallation into
// a Installation record.
func (d DisplayInstallation) ConvertToInstallation() (storage.Installation, error) {
	i := storage.Installation{
		SchemaVersion:  d.SchemaVersion,
		ID:             d.ID,
		Name:           d.Name,
		Namespace:      d.Namespace,
		Uninstalled:    d.Uninstalled,
		Bundle:         d.Bundle,
		Custom:         d.Custom,
		Labels:         d.Labels,
		CredentialSets: d.CredentialSets,
		ParameterSets:  d.ParameterSets,
		Status:         d.Status,
	}

	var err error
	i.Parameters, err = d.ConvertParamToSet(i)
	if err != nil {
		return storage.Installation{}, err
	}

	if err := i.Validate(); err != nil {
		return storage.Installation{}, errors.Wrap(err, "invalid installation")
	}

	return i, nil

}

// ConvertParamToSet converts a Parameters into a internal ParameterSet.
func (d DisplayInstallation) ConvertParamToSet(i storage.Installation) (storage.ParameterSet, error) {
	strategies := make([]secrets.Strategy, 0, len(d.Parameters))
	for name, value := range d.Parameters {
		stringVal, err := cnab.WriteParameterToString(name, value)
		if err != nil {
			return storage.ParameterSet{}, err
		}
		strategy := secrets.Strategy{
			Name:  name,
			Value: stringVal,
		}
		strategies = append(strategies, strategy)
	}

	return storage.NewInternalParameterSet(d.Namespace, d.Name, strategies...), nil
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
	ID         string                 `json:"id" yaml:"id"`
	Bundle     string                 `json:"bundle,omitempty" yaml:"bundle,omitempty"`
	Version    string                 `json:"version" yaml:"version"`
	Action     string                 `json:"action" yaml:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Started    time.Time              `json:"started" yaml:"started"`
	Stopped    *time.Time             `json:"stopped" yaml:"stopped"`
	Status     string                 `json:"status" yaml:"status"`
}

func NewDisplayRun(run storage.Run) DisplayRun {
	return DisplayRun{
		ID:         run.ID,
		Action:     run.Action,
		Parameters: run.TypedParameterValues(),
		Started:    run.Created,
		Bundle:     run.BundleReference,
		Version:    run.Bundle.Version,
	}
}

// ListInstallations lists installed bundles.
func (p *Porter) ListInstallations(ctx context.Context, opts ListOptions) (DisplayInstallations, error) {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	installations, err := p.Installations.ListInstallations(ctx, storage.ListOptions{
		Namespace: opts.GetNamespace(),
		Name:      opts.Name,
		Labels:    opts.ParseLabels(),
		Skip:      opts.Skip,
		Limit:     opts.Limit,
	})
	if err != nil {
		return nil, log.Error(fmt.Errorf("could not list installations: %w", err))
	}

	var displayInstallations DisplayInstallations
	for _, installation := range installations {
		di := NewDisplayInstallation(installation)
		displayInstallations = append(displayInstallations, di)
	}
	sort.Sort(sort.Reverse(displayInstallations))

	return displayInstallations, nil
}

// PrintInstallations prints installed bundles.
func (p *Porter) PrintInstallations(ctx context.Context, opts ListOptions) error {
	displayInstallations, err := p.ListInstallations(ctx, opts)
	if err != nil {
		return err
	}

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
