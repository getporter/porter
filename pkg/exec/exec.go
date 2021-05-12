package exec

import (
	"get.porter.sh/porter/pkg/context"
)

// Exec is the logic behind the exec mixin
type Mixin struct {
	*context.Context
}

// New exec mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context: context.New(),
	}
}
