package porter

import (
	"testing"

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
		},
		TestConfig: c,
	}

	return p
}
