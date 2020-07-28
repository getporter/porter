package claims

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
)

var _ claim.Provider = &ClaimStorage{}

// ClaimStorage provides access to backing claim storage by instantiating
// plugins that implement claim (CRUD) storage.
type ClaimStorage struct {
	*config.Config
	claim.Store
}

func NewClaimStorage(c *config.Config, storagePlugin *pluginstore.Store) *ClaimStorage {
	migration := newMigrateClaimsWrapper(c.Context, storagePlugin)
	return &ClaimStorage{
		Config: c,
		Store:  claim.NewClaimStore(crud.NewBackingStore(migration), nil, nil),
	}
}
