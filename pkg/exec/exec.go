package exec

import (
	"get.porter.sh/porter/pkg/portercontext"
)

// Exec is the logic behind the exec mixin
type Mixin struct {
	*portercontext.Context
}

// New exec mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context: portercontext.New(),
	}
}
