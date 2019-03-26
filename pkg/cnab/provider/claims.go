package cnabprovider

import (
	"github.com/deislabs/duffle/pkg/claim"
	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/utils/crud"
)

func (d *Duffle) NewClaimStore() claim.Store {
	// TODO: I'm going to submit a PR after this so that GetHomeDir doesn't return an error
	homepath, _ := d.GetHomeDir()
	h := home.Home(homepath)
	return claim.NewClaimStore(crud.NewFileSystemStore(h.Claims(), "json"))
}
