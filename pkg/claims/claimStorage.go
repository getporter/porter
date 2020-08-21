package claims

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/claim"
)

var _ claim.Provider = &ClaimStorage{}

// ClaimStorage provides access to backing claim storage by instantiating
// plugins that implement claim (CRUD) storage.
type ClaimStorage struct {
	*config.Config
	claim.Store
}

func NewClaimStorage(storage *storage.Manager) *ClaimStorage {
	return &ClaimStorage{
		Config: storage.Config,
		Store:  claim.NewClaimStore(storage, nil, nil),
	}
}
