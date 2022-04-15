package mongodb_docker

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

var _ plugins.StorageProtocol = &Store{}

// Store is a storage plugin for porter suitable for running on machines
// that have not configured proper storage, i.e. a mongo database.
// It runs mongodb in a docker container and stores its data in a docker volume.
type Store struct {
	plugins.StorageProtocol
	context *portercontext.Context

	config PluginConfig
}

func NewStore(c *portercontext.Context, cfg PluginConfig) *Store {
	s := &Store{
		context: c,
		config:  cfg,
	}

	// This is extra insurance that the db connection is closed
	runtime.SetFinalizer(s, func(s *Store) {
		s.Close(context.Background())
	})

	return s
}

// Connect initializes the plugin for use.
// The plugin itself is responsible for ensuring it was called.
// Close is called automatically when the plugin is used by Porter.
func (s *Store) Connect(ctx context.Context) error {
	if s.StorageProtocol != nil {
		return nil
	}

	// Run mongo in a container storing its data in a volume
	container := "porter-mongodb-docker-plugin"
	dataVol := container + "-data"

	conn, err := EnsureMongoIsRunning(ctx, s.context, container, s.config.Port, dataVol, s.config.Database, s.config.Timeout)
	if err != nil {
		return err
	}

	s.StorageProtocol = conn
	return nil
}

func (s *Store) Close(ctx context.Context) error {
	// leave the container running for performance purposes
	//exec.Command("docker", "rm", "-f", "porter-mongodb-docker-plugin")

	// close the connection to the mongodb running in the container
	if conn, ok := s.StorageProtocol.(*mongodb.Store); ok {
		return conn.Close(ctx)
	}

	s.StorageProtocol = nil

	return nil
}

// EnsureIndex makes sure that the specified index exists as specified.
// If it does exist with a different definition, the index is recreated.
func (s *Store) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.StorageProtocol.EnsureIndex(ctx, opts)
}

func (s *Store) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	return s.StorageProtocol.Aggregate(ctx, opts)
}

func (s *Store) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	if err := s.Connect(ctx); err != nil {
		return 0, err
	}

	return s.StorageProtocol.Count(ctx, opts)
}

func (s *Store) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	return s.StorageProtocol.Find(ctx, opts)
}

func (s *Store) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.StorageProtocol.Insert(ctx, opts)
}

func (s *Store) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.StorageProtocol.Patch(ctx, opts)
}

func (s *Store) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.StorageProtocol.Remove(ctx, opts)
}

func (s *Store) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.StorageProtocol.Update(ctx, opts)
}

func EnsureMongoIsRunning(ctx context.Context, c *portercontext.Context, container string, port string, dataVol string, dbName string, timeoutSeconds int) (*mongodb.Store, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if dataVol != "" {
		err := exec.Command("docker", "volume", "inspect", dataVol).Run()
		if err != nil {
			span.Debugf("creating a data volume, %s, for the mongodb plugin", dataVol)

			err = exec.Command("docker", "volume", "create", dataVol).Run()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					err = fmt.Errorf(string(exitErr.Stderr))
				}
				return nil, span.Error(fmt.Errorf("error creating %s docker volume: %w", dataVol, err))
			}
		}
	}

	// TODO(carolynvs): run this using the docker library
	startMongo := func() error {
		mongoImg := "mongo:4.0-xenial"
		span.Debugf("pulling", mongoImg)

		err := exec.Command("docker", "pull", mongoImg).Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				err = fmt.Errorf(string(exitErr.Stderr))
			}
			return span.Error(fmt.Errorf("error pulling %s: %w", mongoImg, err))
		}

		span.Debugf("running a mongo server in a container on port", port)

		args := []string{"run", "--name", container, "-p=" + port + ":27017", "-d"}
		if dataVol != "" {
			args = append(args, "--mount", "source="+dataVol+",destination=/data/db")
		}
		args = append(args, mongoImg)
		mongoC := exec.Command("docker", args...)
		err = mongoC.Start()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				err = fmt.Errorf(string(exitErr.Stderr))
			}
			return span.Error(fmt.Errorf("error running a mongo container for the mongodb-docker plugin: %w", err))
		}
		return nil
	}
	containerStatus, err := exec.Command("docker", "container", "inspect", container).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && strings.Contains(strings.ToLower(string(exitErr.Stderr)), "no such") { // Container doesn't exist
			if err = startMongo(); err != nil {
				return nil, err
			}
		} else {
			if exitErr, ok := err.(*exec.ExitError); ok {
				err = fmt.Errorf(string(exitErr.Stderr))
			}
			return nil, span.Error(fmt.Errorf("error inspecting container %s: %w", container, err))
		}
	} else if !strings.Contains(string(containerStatus), `"Status": "running"`) { // Container is stopped
		err = exec.Command("docker", "rm", "-f", container).Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				err = fmt.Errorf(string(exitErr.Stderr))
			}
			return nil, span.Error(fmt.Errorf("error cleaning up stopped container %s: %w", container, err))
		}

		if err = startMongo(); err != nil {
			return nil, span.Error(err)
		}
	}

	// wait until the mongo daemon is ready
	span.Debug("waiting for the mongo service to be ready")

	mongoPluginCfg := mongodb.PluginConfig{
		URL:     fmt.Sprintf("mongodb://localhost:%s/%s?connect=direct", port, dbName),
		Timeout: timeoutSeconds,
	}
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	for {
		select {
		case <-timeout.Done():
			return nil, span.Error(errors.New("timeout waiting for local mongodb daemon to be ready"))
		default:
			conn := mongodb.NewStore(c, mongoPluginCfg)
			err := conn.Connect(ctx)
			if err == nil {
				return conn, nil
			}

			time.Sleep(time.Second)
		}
	}
}
