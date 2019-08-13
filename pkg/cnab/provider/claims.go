package cnabprovider

import (
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

const (
	// ClaimsDirectory represents the name of the directory where claims are stored
	ClaimsDirectory = "claims"
)

func (d *Duffle) NewClaimStore() (claim.Store, error) {
	claimsPath, err := d.Config.GetClaimsDir()
	if err != nil {
		return claim.Store{}, errors.Wrap(err, "could not get path to the claims directory")
	}
	return claim.NewClaimStore(crud.NewFileSystemStore(claimsPath, "json")), nil
}

// FetchClaim fetches a claim from the given CNABProvider's claim store
func (d *Duffle) FetchClaim(name string) (*claim.Claim, error) {
	claimStore, err := d.NewClaimStore()
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve claim %s", name)
	}
	claim, err := claimStore.Read(name)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve claim %s", name)
	}
	return &claim, nil
}
