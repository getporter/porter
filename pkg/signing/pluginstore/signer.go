package pluginstore

import (
	"context"
	"errors"
	"fmt"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/signing/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

var _ plugins.SigningProtocol = &Signer{}

// Signer is a plugin-backed source of signing. It resolves the appropriate
// plugin based on Porter's config and implements the plugins.SigningProtocol interface
// using the backing plugin.
//
// Connects just-in-time, but you must call Close to release resources.
type Signer struct {
	*config.Config
	plugin plugins.SigningProtocol
	conn   *pluggable.PluginConnection
}

func NewSigner(c *config.Config) *Signer {
	return &Signer{
		Config: c,
	}
}

// NewSigningPluginConfig for signing sources.
func NewSigningPluginConfig() pluggable.PluginTypeConfig {
	return pluggable.PluginTypeConfig{
		Interface: plugins.PluginInterface,
		Plugin:    &Plugin{},
		GetDefaultPluggable: func(c *config.Config) string {
			return c.Data.DefaultSigning
		},
		GetPluggable: func(c *config.Config, name string) (pluggable.Entry, error) {
			return c.GetSigningPlugin(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			return c.Data.DefaultSigningPlugin
		},
		ProtocolVersion: plugins.PluginProtocolVersion,
	}
}

func (s *Signer) Sign(ctx context.Context, ref string) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("ref", ref))
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	err := s.plugin.Sign(ctx, ref)
	if err != nil {
		return span.Error(err)
	}
	return nil
}

func (s *Signer) Verify(ctx context.Context, ref string) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("ref", ref))
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	err := s.plugin.Verify(ctx, ref)
	if errors.Is(err, plugins.ErrNotImplemented) {
		return span.Error(fmt.Errorf(`the current signing plugin does not support verifying signatures. You need to edit your porter configuration file and configure a different signing plugin: %w`, err))
	}

	return span.Error(err)
}

// Connect initializes the plugin for use.
// The plugin itself is responsible for ensuring it was called.
// Close is called automatically when the plugin is used by Porter.
func (s *Signer) Connect(ctx context.Context) error {
	if s.plugin != nil {
		return nil
	}

	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	pluginType := NewSigningPluginConfig()

	l := pluggable.NewPluginLoader(s.Config)
	conn, err := l.Load(ctx, pluginType)
	if err != nil {
		return span.Error(err)
	}
	s.conn = conn

	store, ok := conn.GetClient().(plugins.SigningProtocol)
	if !ok {
		conn.Close(ctx)
		return span.Error(fmt.Errorf("the interface (%T) exposed by the %s plugin was not plugins.SigningProtocol", conn.GetClient(), conn))
	}
	s.plugin = store

	return nil
}
