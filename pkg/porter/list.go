package porter

import (
	"fmt"
	"sort"
	"time"

	"github.com/cnabio/cnab-go/claim"

	"get.porter.sh/porter/pkg/printer"
	dtprinter "github.com/carolynvs/datetime-printer"
	"github.com/pkg/errors"
)

// ListOptions represent generic options for use by Porter's list commands
type ListOptions struct {
	printer.PrintOptions
}

// DisplayInstallation holds a subset of pertinent values to be listed from a claim.Claim
type DisplayInstallation struct {
	Name     string
	Created  time.Time
	Modified time.Time
	Action   string
	Status   string

	Outputs DisplayOutputs
	History []InstallationAction
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
		history[i] = InstallationAction{
			Action:    hc.Action,
			Timestamp: hc.Created,
			Status:    hc.GetStatus(),
		}
	}

	return DisplayInstallation{
		Name:     installation.Name,
		Created:  installTime,
		Modified: c.Created,
		Action:   c.Action,
		Status:   installation.GetLastStatus(),
		History:  history,
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
	Action    string
	Timestamp time.Time
	Status    string
}

// ListInstallations lists installed bundles by their claims.
func (p *Porter) ListInstallations(opts ListOptions) error {
	installations, err := p.Claims.ReadAllInstallationStatus()
	if err != nil {
		return errors.Wrap(err, "could not list installations")
	}

	var displayInstallations DisplayInstallations
	for _, installation := range installations {
		displayInstallation, err := NewDisplayInstallation(installation)
		if err != nil {
			return err
		}
		displayInstallations = append(displayInstallations, displayInstallation)
	}
	sort.Sort(sort.Reverse(displayInstallations))

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, installations)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, installations)
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
