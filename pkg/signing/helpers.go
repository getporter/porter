package signing

var _ Signer = &TestSigningProvider{}

type TestSigningProvider struct {
	PluginAdapter
	//TODO: add a test signer here
}

func NewTestSigningProvider() TestSigningProvider {
	// TODO: implement this
	return TestSigningProvider{}
}
