package porter

import (
	"encoding/json"
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/mixin"

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

func (p *TestMixinProvider) GetMixinSchema(m mixin.Metadata) (map[string]interface{}, error) {
	t := packr.New("schema", "./schema")

	b, err := t.Find(m.Name + ".json")
	if err != nil {
		return nil, err
	}

	manifestSchema := make(map[string]interface{})
	err = json.Unmarshal(b, &manifestSchema)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal the %s mixin schema", m.Name)
	}

	return manifestSchema, nil
}
