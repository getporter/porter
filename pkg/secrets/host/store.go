package host

import (
	"get.porter.sh/porter/pkg/secrets"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/credentials/host"
)

var _ secrets.Store = &Store{}

// Store implements a store that resolves secrets from the host.
type Store struct {}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Resolve(cred credentials.Source) (string, error) {
	return host.Resolve(cred)
}
