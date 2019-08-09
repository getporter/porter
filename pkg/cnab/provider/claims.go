package cnabprovider

import (
	"path/filepath"

	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/pkg/errors"
)

const (
	// ClaimsDirectory represents the name of the directory where claims are stored
	ClaimsDirectory = "claims"
)

func (d *Duffle) NewClaimStore() claim.Store {
	// TODO: I'm going to submit a PR after this so that GetHomeDir doesn't return an error
	homepath, _ := d.GetHomeDir()
	claimsPath := filepath.Join(homepath, ClaimsDirectory)
	return claim.NewClaimStore(crud.NewFileSystemStore(claimsPath, "json"))
}

// FetchClaim fetches a claim from the given CNABProvider's claim store
func (d *Duffle) FetchClaim(name string) (*claim.Claim, error) {
	claimStore := d.NewClaimStore()
	claim, err := claimStore.Read(name)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve claim %s", name)
	}
	return &claim, nil
}
