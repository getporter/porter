package porter

import (
	"fmt"

	"github.com/pkg/errors"

	cnab "github.com/deislabs/porter/pkg/cnab/provider"
	printer "github.com/deislabs/porter/pkg/printer"
)

// ListOptions represent options for a bundle list command
type ListOptions struct {
	RawFormat string
	Format    printer.Format
}

const (
	// TimeFormat is used to generate a human-readable representation of a raw time.Time value
	TimeFormat = "Mon Jan _2 15:04:05"
)

// CondensedClaim holds a subset of pertinent values to be listed from a claim.Claim
type CondensedClaim struct {
	Name     string
	Created  string
	Modified string
	Action   string
	Status   string
}

// ListBundles lists installed bundles using the printer.Format provided
func (p *Porter) ListBundles(opts printer.PrintOptions) error {
	cp := cnab.NewDuffle(p.Config)
	claimStore := cp.NewClaimStore()
	claims, err := claimStore.ReadAll()
	if err != nil {
		return errors.Wrap(err, "could not list claims")
	}

	var condensedClaims []CondensedClaim
	for _, claim := range claims {
		condensedClaim := CondensedClaim{
			Name:     claim.Name,
			Created:  fmt.Sprint(claim.Created.Format(TimeFormat)),
			Modified: fmt.Sprint(claim.Modified.Format(TimeFormat)),
			Action:   claim.Result.Action,
			Status:   claim.Result.Status,
		}
		condensedClaims = append(condensedClaims, condensedClaim)
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(p.Out, condensedClaims)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, condensedClaims)
	case printer.FormatTable:
		printClaimRow :=
			func(v interface{}) []interface{} {
				cl, ok := v.(CondensedClaim)
				if !ok {
					return nil
				}
				return []interface{}{cl.Name, cl.Created, cl.Modified, cl.Action, cl.Status}
			}
		return printer.PrintTable(p.Out, condensedClaims, printClaimRow,
			"NAME", "CREATED", "MODIFIED", "LAST ACTION", "LAST STATUS")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
