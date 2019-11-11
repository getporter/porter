package filesystem

import (
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
)

var _ crud.Store = &Store{}

// Store is a local filesystem store that stores data in the porter home directory.
type Store struct {
	*storage.DynamicCrudStore
	config.Config
	logger hclog.Logger
}

func NewStore(c config.Config, l hclog.Logger) *Store {
	s := &Store{
		Config: c,
		logger: l,
	}

	s.DynamicCrudStore = storage.NewDynamicCrudStore(s.init)

	return s
}

func (s *Store) init() (crud.Store, func(), error) {
	home, err := s.Config.GetHomeDir()
	if err != nil {
		return nil, nil, err
	}

	store := crud.NewFileSystemStore(home, "json")

	return store, nil, nil
}
