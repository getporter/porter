package host

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/secrets/plugins"
	secretsplugins "get.porter.sh/porter/pkg/secrets/plugins"
	"github.com/cnabio/cnab-go/secrets/host"
)

var _ secretsplugins.SecretsProtocol = Store{}

type Store struct {
	store *host.SecretStore
}

func NewStore() Store {
	return Store{store: &host.SecretStore{}}
}

func (s Store) Resolve(ctx context.Context, keyName string, keyValue string) (string, error) {
	return s.store.Resolve(keyName, keyValue)
}

func (s Store) Create(ctx context.Context, keyName string, keyValue string, value string) error {
	return fmt.Errorf("The default secrets plugin, %s, does not support persisting secrets: %w", PluginKey, plugins.ErrNotImplemented)
}
