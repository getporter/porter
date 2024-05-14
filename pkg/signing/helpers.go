package signing

import "get.porter.sh/porter/pkg/signing/plugins/mock"

var _ Signer = &TestSigningProvider{}

type TestSigningProvider struct {
	PluginAdapter

	signer *mock.Signer
}

func NewTestSigningProvider() TestSigningProvider {
	signer := mock.NewSigner()
	return TestSigningProvider{
		PluginAdapter: NewPluginAdapter(signer),
		signer:        signer,
	}
}
