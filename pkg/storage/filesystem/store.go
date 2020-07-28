package filesystem

import (
	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
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

	home, err := s.Config.GetHomeDir()
	if err != nil {
		return errors.Wrap(err, "could not determine home directory for filesystem storage")
	}

	s.logger.Info("PORTER HOME: " + home)

	s.Store = crud.NewFileSystemStore(home, NewFileExtensions())
	return nil
}

func NewFileExtensions() map[string]string {
	ext := claim.NewClaimStoreFileExtensions()

	// Handle top level files, like schema.json
	ext[""] = ".json"
	return ext
}
