package porter

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"get.porter.sh/porter/pkg/printer"
	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/cnabio/cnab-go/claim"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
)

// ListOptions represent generic options for use by Porter's list commands
type ListOptions struct {
	printer.PrintOptions
}

// DisplayInstallation holds a subset of pertinent values to be listed from installation data
// originating from its claims, results and outputs records
type DisplayInstallation struct {
	Name     string    `json:"name" yaml:"name"`
	Created  time.Time `json:"created" yaml:"created"`
	Modified time.Time `json:"modified" yaml:"modified"`
	Bundle   string    `json:"bundle,omitempty" yaml:"bundle,omitempty"`
	Version  string    `json:"version" yaml:"version"`
	Action   string    `json:"action" yaml:"action"`
	Status   string    `json:"status" yaml:"status"`

	Parameters DisplayValues        `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Outputs    DisplayValues        `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	History    []InstallationAction `json:"history,omitempty" yaml:"history,omitempty"`
}

func NewDisplayInstallation(installation claim.Installation) (DisplayInstallation, error) {
	c, err := installation.GetLastClaim()
	if err != nil {
		return DisplayInstallation{}, err
	}

	installTime, err := installation.GetInstallationTimestamp()
	if err != nil {
		// if we cannot determine when the bundle was installed,
		// for example it hasn't had install run yet, only an action like dry-run
		// just use the timestamp from the claim
		installTime = c.Created
	}

	history := make([]InstallationAction, len(installation.Claims))
	for i, hc := range installation.Claims {
		hasLogs, ok := hc.HasLogs()
		hasLogsS := ""
		if ok {
			hasLogsS = strconv.FormatBool(hasLogs)
		}

		history[i] = InstallationAction{
			ClaimID:    hc.ID,
			Action:     hc.Action,
			Parameters: hc.Parameters,
			Timestamp:  hc.Created,
			Bundle:     tryParseBundleRepository(hc.BundleReference),
			Version:    hc.Bundle.Version,
			Status:     hc.GetStatus(),
			HasLogs:    hasLogsS,
		}
	}

	return DisplayInstallation{
		Name:       installation.Name,
		Bundle:     tryParseBundleRepository(c.BundleReference),
		Version:    c.Bundle.Version,
		Created:    installTime,
		Modified:   c.Created,
		Action:     c.Action,
		Parameters: NewDisplayValuesFromParameters(c.Bundle, c.Parameters),
		Status:     installation.GetLastStatus(),
		History:    history,
	}, nil
}

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

type InstallationAction struct {
	ClaimID    string                 `json:"claimID" yaml:"claimID"`
	Bundle     string                 `json:"bundle,omitempty" yaml:"bundle,omitempty"`
	Version    string                 `json:"version" yaml:"version"`
	Action     string                 `json:"action" yaml:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Timestamp  time.Time              `json:"timestamp" yaml:"timestamp"`
	Status     string                 `json:"status" yaml:"status"`
	HasLogs    string                 `json:"hasLogs" yaml:"hasLogs"`
}

// ListInstallations lists installed bundles.
func (p *Porter) ListInstallations() (DisplayInstallations, error) {
	installations, err := p.Claims.ReadAllInstallationStatus()
	if err != nil {
		return nil, errors.Wrap(err, "could not list installations")
	}

	var displayInstallations DisplayInstallations
	for _, installation := range installations {
		displayInstallation, err := NewDisplayInstallation(installation)
		if err != nil {
			return nil, err
		}
		displayInstallations = append(displayInstallations, displayInstallation)
	}
	sort.Sort(sort.Reverse(displayInstallations))

	return displayInstallations, nil
}

// PrintInstallations prints installed bundles.
func (p *Porter) PrintInstallations(opts ListOptions) error {
	displayInstallations, err := p.ListInstallations()
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, displayInstallations)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, displayInstallations)
	case printer.FormatTable:
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
				return []interface{}{cl.Name, tp.Format(cl.Created), tp.Format(cl.Modified), cl.Action, cl.Status}
			}
		return printer.PrintTable(p.Out, displayInstallations, row,
			"NAME", "CREATED", "MODIFIED", "LAST ACTION", "LAST STATUS")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func tryParseBundleRepository(bundleReference string) string {
	if ref, err := reference.ParseNormalizedNamed(bundleReference); err == nil {
		return reference.FamiliarName(ref)
	}
	return bundleReference
}
