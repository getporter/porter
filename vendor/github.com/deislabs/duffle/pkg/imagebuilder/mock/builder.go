package mock

import (
	"context"
	"io"

	"github.com/deislabs/duffle/pkg/duffle/manifest"
)

// Builder represents a mock builder
type Builder struct {
}

// Name represents the name of an image the mock builder will build
func (b Builder) Name() string {
	return "cnab"
}

// Type represents the type of a mock builder
func (b Builder) Type() string {
	return "mock-type"
}

// URI represents the URI of the artefact of a mock builder
func (b Builder) URI() string {
	return "mock-uri:1.0.0"
}

// Digest represents the digest of a mock builder
func (b Builder) Digest() string {
	return "mock-digest"
}

// NewBuilder returns a new mock builder
func NewBuilder(c *manifest.InvocationImage) *Builder {
	return &Builder{}
}

// PrepareBuild is no-op for a mock builder
func (b *Builder) PrepareBuild(appDir, registry, name string) error {
	return nil
}

// Build is no-op for a mock builder
func (b Builder) Build(ctx context.Context, log io.WriteCloser) error {
	return nil
}
