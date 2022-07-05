package migrations

import (
	"context"
	"fmt"
	"path/filepath"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/storage/migrations/crudstore"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
)

// LegacyPluginAdapter can connect to a legacy Porter v0.38 storage plugin to
// retrieve data to migrate to the current version of Porter.
type LegacyPluginAdapter struct {
	config *config.Config

	// The name of the source storage account from which we will read old Porter data.
	storageName string

	// The path to an older PORTER_HOME directory containing the previous version
	// of Porter and compatible plugins.
	oldPorterHome string

	// The legacy porter v0.38 storage interface that we will use with the old plugins
	// to retrieve data.
	store crudstore.Store

	// Connection to the legacy plugin
	pluginConn *pluggable.PluginConnection
}

func NewLegacyPluginAdapter(c *config.Config, oldPorterHome string, storageName string) *LegacyPluginAdapter {
	return &LegacyPluginAdapter{
		config:        c,
		oldPorterHome: oldPorterHome,
		storageName:   storageName,
	}
}

// Connect loads the legacy plugin specified by the source storage account.
func (m *LegacyPluginAdapter) Connect(ctx context.Context) error {
	ctx, log := tracing.StartSpan(ctx,
		attribute.String("storage-name", m.storageName))
	defer log.EndSpan()

	// Create a config file that uses the old PORTER_HOME
	oldConfig := config.New()
	oldConfig.SetHomeDir(m.oldPorterHome)
	oldConfig.SetPorterPath(filepath.Join(m.oldPorterHome, "porter"))
	oldConfig.Load(ctx, func(ctx context.Context, secretKey string) (string, error) {
		return "", nil
	})
	oldConfig.Setenv(config.EnvHOME, m.oldPorterHome)

	// Start the plugin
	l := pluggable.NewPluginLoader(oldConfig)
	conn, err := l.Load(ctx, m.makePluginConfig())
	if err != nil {
		return log.Error(fmt.Errorf("could not load legacy storage plugin: %w", err))
	}
	m.pluginConn = conn

	// Ensure we close the plugin connection if anything fails
	// Afterwards its on the caller to close it
	connected := false
	defer func() {
		if !connected {
			conn.Close(ctx)
		}
	}()

	// Cast the plugin connection to a subset of the old protocol from v0.38 that can only read data
	store, ok := conn.GetClient().(crudstore.Store)
	if !ok {
		return log.Error(fmt.Errorf("the interface exposed by the %s plugin was not crudstore.Store", conn))
	}

	m.store = store
	connected = true
	return nil
}

// makePluginConfig creates a plugin configuration compatible with the legacy plugins
// from porter v0.38
func (m *LegacyPluginAdapter) makePluginConfig() pluggable.PluginTypeConfig {
	return pluggable.PluginTypeConfig{
		Interface: plugins.PluginInterface,
		Plugin:    &crudstore.Plugin{},
		GetDefaultPluggable: func(c *config.Config) string {
			// Load the config for the specific storage account named as the source for the migration
			return m.storageName
		},
		GetPluggable: func(c *config.Config, name string) (pluggable.Entry, error) {
			return c.GetStorage(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			// filesystem is the default storage plugin for v0.38
			return "filesystem"
		},
		ProtocolVersion: 1, // protocol version used by porter v0.38
	}
}

func (m *LegacyPluginAdapter) Close() error {
	if m.pluginConn != nil {
		m.pluginConn.Close(context.Background())
		m.pluginConn = nil
	}
	return nil
}

func (m *LegacyPluginAdapter) List(ctx context.Context, itemType string, group string) ([]string, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}

	return m.store.List(itemType, group)
}

func (m *LegacyPluginAdapter) Read(ctx context.Context, itemType string, name string) ([]byte, error) {
	if err := m.Connect(ctx); err != nil {
		return nil, err
	}

	return m.store.Read(itemType, name)
}
