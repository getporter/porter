package runtime

import (
	"context"
	"strconv"

	"get.porter.sh/porter/pkg/config"
)

// RuntimeConfig is a specialized config.Config with additional runtime-specific settings.
type RuntimeConfig struct {
	*config.Config

	// DebugMode indicates if the bundle is running in debug mode.
	DebugMode bool
}

// NewConfig returns an initialized RuntimeConfig
func NewConfig() RuntimeConfig {
	return NewConfigFor(config.New())
}

// NewConfigFor returns an initialized RuntimeConfig using the specified context.
func NewConfigFor(config *config.Config) RuntimeConfig {
	debug, _ := strconv.ParseBool(config.Getenv("PORTER_DEBUG"))
	return RuntimeConfig{
		DebugMode: debug,
		Config:    config,
	}
}

func (c RuntimeConfig) ConfigureLogging(ctx context.Context) (context.Context, error) {
	// Just use porter's config to load up common settings, such as logging
	pc := config.NewFor(c.Context)
	return pc.Load(ctx, nil)
}
