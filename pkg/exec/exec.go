//go:generate packr2

package exec

import (
	"get.porter.sh/porter/pkg/context"
	"github.com/gobuffalo/packr/v2"
)

// Exec is the logic behind the exec mixin
type Mixin struct {
	*context.Context

	schemas *packr.Box
}

// New exec mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context: context.New(),
		schemas: NewSchemaBox(),
	}
}

func NewSchemaBox() *packr.Box {
	return packr.New("get.porter.sh/porter/pkg/exec/schema", "./schema")
}
