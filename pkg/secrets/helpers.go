package secrets

import inmemory "get.porter.sh/porter/pkg/secrets/plugins/in-memory"

var _ Store = &TestSecretsProvider{}

type TestSecretsProvider struct {
	PluginAdapter

	secrets *inmemory.Store
}

func NewTestSecretsProvider() TestSecretsProvider {
	secrets := inmemory.NewStore()
	return TestSecretsProvider{
		PluginAdapter: NewPluginAdapter(secrets),
		secrets:       secrets,
	}
}

func (s TestSecretsProvider) Close() error {
	return nil
}
