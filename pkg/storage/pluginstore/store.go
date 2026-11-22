package pluginstore

import (
	"context"
	"fmt"
	"io"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/plugins/pluggable"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"get.porter.sh/porter/pkg/tracing"
)

var _ io.Closer = &Store{}
var _ plugins.StorageBackend = &Store{}

// Store is a plugin-backed source of storage. It resolves the appropriate
// plugin based on Porter's config and implements the plugins.StorageBackend
// interface using the backing plugin.
//
// Connects just-in-time, but you must call Close to release resources.
type Store struct {
	*config.Config
	backend plugins.StorageBackend
	conn    *pluggable.PluginConnection
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
		Plugin:    &PluginV4{},
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
	if s.backend != nil {
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

	backend, ok := conn.GetClient().(plugins.StorageBackend)
	if !ok {
		conn.Close(ctx)
		s.conn = nil
		return span.Error(fmt.Errorf("the interface (%T) exposed by the %s plugin was not plugins.StorageBackend", conn.GetClient(), conn))
	}

	s.backend = backend

	return nil
}

func (s *Store) Close() error {
	if s.conn != nil {
		s.conn.Close(context.Background())
		s.conn = nil
	}
	s.backend = nil
	return nil
}

func (s *Store) Info(ctx context.Context, req *protov4.InfoRequest) (*protov4.InfoResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Info(ctx, req)
}

func (s *Store) Migrate(ctx context.Context, req *protov4.MigrateRequest) (*protov4.MigrateResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Migrate(ctx, req)
}

func (s *Store) Put(ctx context.Context, req *protov4.PutRequest) (*protov4.PutResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Put(ctx, req)
}

func (s *Store) Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Get(ctx, req)
}

func (s *Store) List(ctx context.Context, req *protov4.ListRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.List(ctx, req)
}

func (s *Store) Remove(ctx context.Context, req *protov4.RemoveRequest) (*protov4.RemoveResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.Remove(ctx, req)
}

func (s *Store) RemoveInstallation(ctx context.Context, req *protov4.RemoveInstallationRequest) (*protov4.RemoveInstallationResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.RemoveInstallation(ctx, req)
}

func (s *Store) ListRuns(ctx context.Context, req *protov4.ListRunsRequest) (*protov4.ListRunsResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.ListRuns(ctx, req)
}

func (s *Store) GetLastRun(ctx context.Context, req *protov4.InstallationRef) (*protov4.GetResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.GetLastRun(ctx, req)
}

func (s *Store) GetActiveRuns(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.GetActiveRuns(ctx, req)
}

func (s *Store) ListResults(ctx context.Context, req *protov4.ListResultsRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.ListResults(ctx, req)
}

func (s *Store) InsertOutputs(ctx context.Context, req *protov4.InsertOutputsRequest) (*protov4.InsertOutputsResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.InsertOutputs(ctx, req)
}

func (s *Store) ListOutputs(ctx context.Context, req *protov4.ListOutputsRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.ListOutputs(ctx, req)
}

func (s *Store) GetOutput(ctx context.Context, req *protov4.GetOutputRequest) (*protov4.GetResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.GetOutput(ctx, req)
}

func (s *Store) GetLastOutputs(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.GetLastOutputs(ctx, req)
}

func (s *Store) FindInstallationsReferencing(ctx context.Context, req *protov4.FindReferencingRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.FindInstallationsReferencing(ctx, req)
}

func (s *Store) FindDependencyCandidates(ctx context.Context, req *protov4.FindDependencyCandidatesRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return s.backend.FindDependencyCandidates(ctx, req)
}
