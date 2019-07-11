package cnabprovider

import (
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/pkg/errors"
)

func (d *Duffle) NewClaimStore() claim.Store {
	// TODO: I'm going to submit a PR after this so that GetHomeDir doesn't return an error
	homepath, _ := d.GetHomeDir()
	h := home.Home(homepath)
	return claim.NewClaimStore(crud.NewFileSystemStore(h.Claims(), "json"))
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
