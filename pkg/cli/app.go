package cli

import (
	"context"

	"get.porter.sh/porter/pkg/config"
)

// PorterApp represents the application that goes with a cobra command.
type PorterApp interface {
	// Connect performs any startup initialization required by the application.
	Connect(ctx context.Context) error

	// Close releases any resources held by your application.
	// Any errors returned are logged.
	Close() error

	// GetConfig returns the Porter configuration for your application.
	GetConfig() *config.Config
}
