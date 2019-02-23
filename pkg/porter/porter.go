//go:generate packr2

package porter

import (
	"github.com/deislabs/porter/pkg/config"
	mixinprovider "github.com/deislabs/porter/pkg/mixin/provider"
	"github.com/gobuffalo/packr/v2"
)

// Porter is the logic behind the porter client.
type Porter struct {
	*config.Config
	MixinProvider

	schemas *packr.Box
}

// New porter client, initialized with useful defaults.
func New() *Porter {
	c := config.New()
	return &Porter{
		Config:        c,
		MixinProvider: mixinprovider.NewFileSystem(c),
		schemas:       NewSchemaBox(),
	}
}

// NewSchemas creates or retrieves the packr box with the porter schemas files.
func NewSchemaBox() *packr.Box {
	return packr.New("github.com/deislabs/porter/pkg/porter/schema", "./schema")
}
