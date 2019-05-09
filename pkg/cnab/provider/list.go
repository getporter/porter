package cnabprovider

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	printer "github.com/deislabs/porter/pkg/printer"
)

// CondensedClaim holds a subset of pertinent values to be listed from a claim.Claim
type CondensedClaim struct {
	Name    string
	Created time.Time
	Action  string
	Status  string
}

// List lists bundles with the printer.Format provided
func (d *Duffle) List(opts printer.PrintOptions) error {
	claimStore := d.NewClaimStore()
	claims, err := claimStore.ReadAll()
	if err != nil {
		return errors.Wrap(err, "could not list claims")
	}

	var condensedClaims []CondensedClaim
	for _, claim := range claims {
		condensedClaim := CondensedClaim{
			Name:    claim.Name,
			Created: claim.Created,
			Action:  claim.Result.Action,
			Status:  claim.Result.Status,
		}
		condensedClaims = append(condensedClaims, condensedClaim)
	}

	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(d.Out, condensedClaims)
	case printer.FormatYaml:
		return printer.PrintYaml(d.Out, condensedClaims)
	case printer.FormatTable:
		printClaimRow :=
			func(v interface{}) []interface{} {
				cl, ok := v.(CondensedClaim)
				if !ok {
					return nil
				}
				return []interface{}{cl.Name, cl.Created, cl.Action, cl.Status}
			}
		return printer.PrintTable(d.Out, condensedClaims, printClaimRow, "NAME", "INSTALLED", "LAST ACTION", "LAST STATUS")
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}
