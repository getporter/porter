package filesystem

import (
	"path/filepath"

	instancestorage "get.porter.sh/porter/pkg/instance-storage"

	"github.com/hashicorp/go-hclog"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/utils/crud"
)

var _ crud.Store = &Store{}

// Store is a local filesystem store that stores claims in the porter home directory.
type Store struct {
	*instancestorage.DynamicCrudStore
	config.Config
	logger hclog.Logger
}

func NewStore(c config.Config, l hclog.Logger) *Store {
	s := &Store{
		Config: c,
		logger: l,
	}

	s.DynamicCrudStore = instancestorage.NewDynamicCrudStore(s.init)

	return s
}

func (s *Store) init() (crud.Store, func(), error) {
	home, err := s.Config.GetHomeDir()
	if err != nil {
		return nil, nil, err
	}

	claimsPath := filepath.Join(home, "claims")
	store := crud.NewFileSystemStore(claimsPath, "json")

	return store, nil, nil
}
