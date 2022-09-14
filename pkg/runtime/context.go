package runtime

import (
	"context"
	"strconv"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
)

// RuntimeConfig is a specialized portercontext.Context with additional runtime-specific settings.
type RuntimeConfig struct {
	*portercontext.Context

	// DebugMode indicates if the bundle is running in debug mode.
	DebugMode bool
}

// NewConfig returns an initialized RuntimeConfig
func NewConfig() RuntimeConfig {
	return NewConfigFor(portercontext.New())
}

// NewConfigFor returns an initialized RuntimeConfig using the specified context.
func NewConfigFor(porterCtx *portercontext.Context) RuntimeConfig {
	debug, _ := strconv.ParseBool(porterCtx.Getenv("PORTER_DEBUG"))
	return RuntimeConfig{
		Context:   porterCtx,
		DebugMode: debug,
	}
}
	}
}
