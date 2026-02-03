package pluginstore

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/sbom/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

var _ plugins.SBOMGeneratorProtocol = &SBOMGenerator{}

// SBOMGenerator is a plugin-backed source of signing. It resolves the appropriate
// plugin based on Porter's config and implements the plugins.SBOMGeneratorProtocol interface
// using the backing plugin.
//
// Connects just-in-time, but you must call Close to release resources.
type SBOMGenerator struct {
	*config.Config
	plugin plugins.SBOMGeneratorProtocol
	conn   *pluggable.PluginConnection
}

func NewSBOMGenerator(c *config.Config) *SBOMGenerator {
	return &SBOMGenerator{
		Config: c,
	}
}

// NewSigningPluginConfig for signing sources.
func NewSigningPluginConfig() pluggable.PluginTypeConfig {
	return pluggable.PluginTypeConfig{
		Interface: plugins.PluginInterface,
		Plugin:    &Plugin{},
		GetDefaultPluggable: func(c *config.Config) string {
			return c.Data.DefaultSBOMGenerator
		},
		GetPluggable: func(c *config.Config, name string) (pluggable.Entry, error) {
			return c.GetSBOMGeneratorPlugin(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			return c.Data.DefaultSBOMGeneratorPlugin
		},
		ProtocolVersion: plugins.PluginProtocolVersion,
	}
}

func (s *SBOMGenerator) Generate(ctx context.Context, ref string, filePath string, insecureRegistry bool) error {
	ctx, span := tracing.StartSpan(ctx,
		attribute.String("ref", ref))
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	err := s.plugin.Generate(ctx, ref, filePath, insecureRegistry)
	if err != nil {
		return span.Error(err)
	}
	return nil
}

// Connect initializes the plugin for use.
// The plugin itself is responsible for ensuring it was called.
// Close is called automatically when the plugin is used by Porter.
func (s *SBOMGenerator) Connect(ctx context.Context) error {
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

	store, ok := conn.GetClient().(plugins.SBOMGeneratorProtocol)
	if !ok {
		conn.Close(ctx)
		return span.Error(fmt.Errorf("the interface (%T) exposed by the %s plugin was not plugins.SBOMGeneratorProtocol", conn.GetClient(), conn))
	}
	s.plugin = store

	if err = s.plugin.Connect(ctx); err != nil {
		return err
	}

	return nil
}

func (s *SBOMGenerator) Close() error {
	if s.conn != nil {
		s.conn.Close(context.Background())
		s.conn = nil
	}
	return nil
}
