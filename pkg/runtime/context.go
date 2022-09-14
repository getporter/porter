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

func (c RuntimeConfig) ConfigureLogging(ctx context.Context) error {
	// Just use porter's config to load up common settings, such as logging
	pc := config.NewFor(c.Context)
	if err := pc.Load(ctx, nil); err != nil {
		return err
	}

	opts := pc.NewLogConfiguration()
	c.Context.ConfigureLogging(ctx, opts)

	return nil
}
