package exec

import (
	"get.porter.sh/porter/pkg/runtime"
)

// Mixin is the logic behind the exec mixin
type Mixin struct {
	// Config is a specialized portercontext.Context with additional runtime settings.
	Config runtime.RuntimeConfig

	// Debug specifies if the mixin should be in debug mode
	Debug bool
}

// New exec mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Config: runtime.NewConfig(),
	}
}
