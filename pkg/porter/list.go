package porter

import (
	"fmt"
	"sort"
	"time"

	dtprinter "github.com/carolynvs/datetime-printer"
	cnab "github.com/deislabs/porter/pkg/cnab/provider"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/pkg/errors"
)

// ListOptions represent generic options for use by Porter's list commands
type ListOptions struct {
	printer.PrintOptions
}

// CondensedClaim holds a subset of pertinent values to be listed from a claim.Claim
type CondensedClaim struct {
	Name     string
	Created  time.Time
	Modified time.Time
	Action   string
	Status   string
}

type CondensedClaimList []CondensedClaim

func (l CondensedClaimList) Len() int {
	return len(l)
}
func (l CondensedClaimList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l CondensedClaimList) Less(i, j int) bool {
	return l[i].Modified.Before(l[j].Modified)
}

// ListInstances lists installed bundles by their claims.
func (p *Porter) ListInstances(opts ListOptions) error {
	cp := cnab.NewDuffle(p.Config)
	claimStore, err := cp.NewClaimStore()
	if err != nil {
		return errors.Wrapf(err, "could not access claim store")
	}
	claims, err := claimStore.ReadAll()
	if err != nil {
		return errors.Wrap(err, "could not list bundle instances")
	}

	var condensedClaims CondensedClaimList
	for _, claim := range claims {
		condensedClaim := CondensedClaim{
			Name:     claim.Name,
			Created:  claim.Created,
			Modified: claim.Modified,
			Action:   claim.Result.Action,
			Status:   claim.Result.Status,
		}
		condensedClaims = append(condensedClaims, condensedClaim)
	}
	sort.Sort(sort.Reverse(condensedClaims))

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, condensedClaims)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, condensedClaims)
	case printer.FormatTable:
		// have every row use the same "now" starting ... NOW!
		now := time.Now()
		tp := dtprinter.DateTimePrinter{
			Now: func() time.Time { return now },
		}

		printClaimRow :=
			func(v interface{}) []interface{} {
				cl, ok := v.(CondensedClaim)
				if !ok {
					return nil
				}
				return []interface{}{cl.Name, tp.Format(cl.Created), tp.Format(cl.Modified), cl.Action, cl.Status}
			}
		return printer.PrintTable(p.Out, condensedClaims, printClaimRow,
			"NAME", "CREATED", "MODIFIED", "LAST ACTION", "LAST STATUS")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
