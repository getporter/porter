package pluginstore

import (
	"context"
	"fmt"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/tracing"
	"go.mongodb.org/mongo-driver/bson"
)

var _ plugins.StorageProtocol = &Store{}

// Store is a plugin-backed source of storage. It resolves the appropriate
// plugin based on Porter's config and implements the plugins.StorageProtocol interface
// using the backing plugin.
//
// Connects just-in-time, but you must call Close to release resources.
type Store struct {
	*config.Config
	plugin plugins.StorageProtocol
	conn   pluggable.PluginConnection
}

func NewStore(c *config.Config) *Store {
	return &Store{
		Config: c,
	}
}

// NewStoragePluginConfig for porter home storage.
func NewStoragePluginConfig() pluggable.PluginTypeConfig {
	return pluggable.PluginTypeConfig{
		Interface: plugins.PluginInterface,
		Plugin:    &Plugin{},
		GetDefaultPluggable: func(c *config.Config) string {
			return c.Data.DefaultStorage
		},
		GetPluggable: func(c *config.Config, name string) (pluggable.Entry, error) {
			return c.GetStorage(name)
		},
		GetDefaultPlugin: func(c *config.Config) string {
			return c.Data.DefaultStoragePlugin
		},
		ProtocolVersion: plugins.PluginProtocolVersion,
	}
}

// Connect initializes the plugin for use.
// The plugin itself is responsible for ensuring it was called.
// Close is called automatically when the plugin is used by Porter.
func (s *Store) Connect(ctx context.Context) error {
	if s.plugin != nil {
		return nil
	}

	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	pluginType := NewStoragePluginConfig()

	l := pluggable.NewPluginLoader(s.Config)
	conn, err := l.Load(ctx, pluginType)
	if err != nil {
		return span.Error(fmt.Errorf("could not load %s plugin: %w", pluginType.Interface, err))
	}
	s.conn = conn

	store, ok := conn.Client.(plugins.StorageProtocol)
	if !ok {
		conn.Close()
		return span.Error(fmt.Errorf("the interface exposed by the %s plugin was not plugins.StorageProtocol", l.SelectedPluginKey))
	}

	s.plugin = store

	return nil
}

func (s *Store) Close(ctx context.Context) error {
	s.conn.Close()
	s.plugin = nil
	return nil
}

func (s *Store) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	err := s.plugin.EnsureIndex(ctx, opts)
	return span.Error(err)
}

func (s *Store) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	results, err := s.plugin.Aggregate(ctx, opts)
	if err != nil {
		return nil, span.Error(err)
	}

	return results, nil
}

func (s *Store) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return 0, err
	}

	count, err := s.plugin.Count(ctx, opts)
	if err != nil {
		return 0, span.Error(err)
	}
	return count, nil
}

func (s *Store) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	results, err := s.plugin.Find(ctx, opts)
	return results, span.Error(err)
}

func (s *Store) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	err := s.plugin.Insert(ctx, opts)
	return span.Error(err)
}

func (s *Store) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	err := s.plugin.Patch(ctx, opts)
	return span.Error(err)
}

func (s *Store) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	err := s.plugin.Remove(ctx, opts)
	return span.Error(err)
}

func (s *Store) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return err
	}

	err := s.plugin.Update(ctx, opts)
	return span.Error(err)
}
