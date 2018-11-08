package porter

import (
	"github.com/deislabs/porter/pkg/config"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	return &Porter{
		Config: config.New(),
	}
}
