package claims

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/pluginstore"
	"github.com/cnabio/cnab-go/claim"
)

var _ ClaimProvider = &ClaimStorage{}

// ClaimStorage provides access to backing claim storage by instantiating
// plugins that implement claim (CRUD) storage.
type ClaimStorage struct {
	*config.Config
	claim.Store
}

func NewClaimStorage(c *config.Config, storagePlugin *pluginstore.Store) *ClaimStorage {
	return &ClaimStorage{
		Config: c,
		Store:  claim.NewClaimStore(storagePlugin),
	}
}
