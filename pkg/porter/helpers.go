package porter

import (
	"testing"

	"github.com/deislabs/porter/pkg/mixin"
	mixinprovider "github.com/deislabs/porter/pkg/mixin/provider"
	"github.com/gobuffalo/packr/v2"

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
			Config: c.Config,
			MixinProvider: &TestMixinProvider{
				MixinProvider: mixinprovider.NewFileSystem(c.Config),
			},
		},
		TestConfig: c,
	}

	return p
}

// TODO: use this later to not actually execute a mixin during a unit test
type TestMixinProvider struct {
	MixinProvider
}

func (p *TestMixinProvider) GetMixins() ([]mixin.Metadata, error) {
	mixins := []mixin.Metadata{
		{Name: "exec"},
	}
	return mixins, nil
}

func (p *TestMixinProvider) GetMixinSchema(m mixin.Metadata) (string, error) {
	t := packr.New("schema", "./schema")

	return t.FindString(m.Name + ".json")
}
