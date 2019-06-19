package cnabprovider

import (
	"github.com/deislabs/cnab-go/claim"
	"github.com/deislabs/cnab-go/utils/crud"
	"github.com/deislabs/duffle/pkg/duffle/home"
)

func (d *Duffle) NewClaimStore() claim.Store {
	// TODO: I'm going to submit a PR after this so that GetHomeDir doesn't return an error
	homepath, _ := d.GetHomeDir()
	h := home.Home(homepath)
	return claim.NewClaimStore(crud.NewFileSystemStore(h.Claims(), "json"))
}
