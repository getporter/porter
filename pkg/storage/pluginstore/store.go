package pluginstore

import (
	"context"

	"go.uber.org/zap/zapcore"

	"get.porter.sh/porter/pkg/config"
	porterplugins "get.porter.sh/porter/pkg/plugins"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

var _ plugins.StoragePlugin = &Store{}

// Store is a plugin-backed source of storage. It resolves the appropriate
// plugin based on Porter's config and implements the plugins.StorageProtocol interface
// using the backing plugin.
//
// Connects just-in-time, but you must call Close to release resources.
type Store struct {
	*config.Config
	plugin plugins.StorageProtocol
	conn   pluggable.PluginConnection
	tracer tracing.Tracer
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
		ProtocolVersion: 2,
	}
}

func CreateInternalPlugin(ctx context.Context, c *portercontext.Context, key string, pluginConfig interface{}) (porterplugins.Plugin, error) {
	switch key {
	case mongodb_docker.PluginKey:
		return mongodb_docker.NewPlugin(ctx, c, pluginConfig)
	case mongodb.PluginKey:
		return mongodb.NewPlugin(ctx, c, pluginConfig)
	default:
		return nil, errors.Errorf("unsupported internal storage plugin specified %s", key)
	}
}

func (s *Store) Connect(ctx context.Context) error {
	s.plugin.SetCommandContext(ctx)
	if s.plugin != nil {
		return nil
	}

	pluginType := NewStoragePluginConfig()

	l := pluggable.NewPluginLoader(s.Config, func(ctx context.Context, key string, config interface{}) (porterplugins.Plugin, error) {
		return CreateInternalPlugin(ctx, s.Context, key, config)
	})
	conn, err := l.Load(ctx, pluginType)
	if err != nil {
		return errors.Wrapf(err, "could not load %s plugin", pluginType.Interface)
	}
	s.conn = conn

	store, ok := conn.Client.(plugins.StorageProtocol)
	if !ok {
		conn.Close()
		return errors.Errorf("the interface exposed by the %s plugin was not plugins.StorageProtocol", l.SelectedPluginKey)
	}
	s.plugin = store
	s.tracer = s.NewTracer(ctx, conn.Key)

	return nil
}

func (s *Store) Close(ctx context.Context) error {
	s.conn.Close()
	s.plugin = nil
	return nil
}

func (s *Store) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.plugin.EnsureIndex(opts)
}

func (s *Store) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	return s.plugin.Aggregate(opts)
}

func (s *Store) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	if err := s.Connect(ctx); err != nil {
		return 0, err
	}

	return s.plugin.Count(opts)
}

func (s *Store) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	ctx, span := tracing.StartSpanForComponent(ctx, s.tracer)
	defer s.attachLogs(span)
	defer span.EndSpan()

	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	return s.plugin.Find(opts)
}

// try to associate logs from the plugin with a particular operation
func (s *Store) attachLogs(span tracing.TraceLogger) {
	for {
		select {
		// Read all available logs and associate them with the current span
		case evt := <-s.conn.Logs:
			switch evt.Level {
			case zapcore.ErrorLevel:
				span.Errorf(evt.Message)
			case zapcore.WarnLevel:
				span.Warn(evt.Message)
			case zapcore.InfoLevel:
				span.Info(evt.Message)
			default:
				span.Debug(evt.Message)
			}
		default:
			// Stop reading as soon as we would need to block to get the next log
			return
		}
	}
}

func (s *Store) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.plugin.Insert(opts)
}

func (s *Store) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.plugin.Patch(opts)
}

func (s *Store) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.plugin.Remove(opts)
}

func (s *Store) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.plugin.Update(opts)
}
