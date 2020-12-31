package filesystem

import (
	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
)

var _ crud.Store = &Store{}

// Store is a local filesystem store that stores data in the porter home directory.
type Store struct {
	crud.Store
	config.Config
	logger hclog.Logger
}

func NewStore(c config.Config, l hclog.Logger) crud.Store {
	// Wrapping ourselves in a backing store so that our Connect is used.
	return crud.NewBackingStore(&Store{
		Config: c,
		logger: l,
	})
}

func (s *Store) Connect() error {
	if s.Store != nil {
		return nil
	}

	home := s.GetHomeDir()
	s.logger.Info("PORTER HOME: " + home)

	s.Store = crud.NewFileSystemStore(home, NewFileExtensions())
	return nil
}

func NewFileExtensions() map[string]string {
	ext := claim.NewClaimStoreFileExtensions()

	jsonExt := ".json"
	ext[credentials.ItemType] = jsonExt

	// TODO (carolynvs): change to parameters.ItemType once parameters move to cnab-go
	ext["parameters"] = jsonExt

	// Handle top level files, like schema.json
	ext[""] = jsonExt

	return ext
}
