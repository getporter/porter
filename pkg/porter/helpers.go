package porter

import (
	"testing"

	mixinprovider "github.com/deislabs/porter/pkg/mixin/provider"

	"github.com/deislabs/porter/pkg/config"
)

type TestPorter struct {
	*Porter
	TestConfig *config.TestConfig
}

// NewTestPorter initializes a porter test client, with the output buffered, and an in-memory file system.
func NewTestPorter(t *testing.T) *TestPorter {
	c := config.NewTestConfig(t)
	p := &TestPorter{
		Porter: &Porter{
			Config:        c.Config,
			MixinProvider: mixinprovider.NewFileSystem(c.Config),
		},
		TestConfig: c,
	}

	return p
}

// TODO: use this later to not actually execute a mixin during a unit test
type TestMixinProvider struct {
	*mixinprovider.FileSystem
}
