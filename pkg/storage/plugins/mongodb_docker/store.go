package mongodb_docker

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb"
	"get.porter.sh/porter/pkg/tracing"
	"go.mongodb.org/mongo-driver/bson"
)

var _ plugins.StorageProtocol = &Store{}

// Store is a storage plugin for porter suitable for running on machines
// that have not configured proper storage, i.e. a mongo database.
// It runs mongodb in a docker container and stores its data in a docker volume.
type Store struct {
	*mongodb.Store
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
		s.Close()
	})

	return s
}

// Connect initializes the plugin for use.
// The plugin itself is responsible for ensuring it was called.
// Close is called automatically when the plugin is used by Porter.
func (s *Store) Connect(ctx context.Context) error {
	if s.Store != nil {
		return nil
	}

	// Run mongo in a container storing its data in a volume
	container := "porter-mongodb-docker-plugin"
	dataVol := container + "-data"

	conn, err := EnsureMongoIsRunning(ctx, s.context, container, s.config.Port, dataVol, s.config.Database, s.config.Timeout)
	if err != nil {
		return err
	}

	s.Store = conn
	return nil
}

func (s *Store) Close() error {
	if s.Store == nil {
		return nil
	}

	err := s.Store.Close()
	s.Store = nil
	return err
}

// EnsureIndex makes sure that the specified index exists as specified.
// If it does exist with a different definition, the index is recreated.
func (s *Store) EnsureIndex(ctx context.Context, opts plugins.EnsureIndexOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.Store.EnsureIndex(ctx, opts)
}

func (s *Store) Aggregate(ctx context.Context, opts plugins.AggregateOptions) ([]bson.Raw, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	return s.Store.Aggregate(ctx, opts)
}

func (s *Store) Count(ctx context.Context, opts plugins.CountOptions) (int64, error) {
	if err := s.Connect(ctx); err != nil {
		return 0, err
	}

	return s.Store.Count(ctx, opts)
}

func (s *Store) Find(ctx context.Context, opts plugins.FindOptions) ([]bson.Raw, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}

	return s.Store.Find(ctx, opts)
}

func (s *Store) Insert(ctx context.Context, opts plugins.InsertOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.Store.Insert(ctx, opts)
}

func (s *Store) Patch(ctx context.Context, opts plugins.PatchOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.Store.Patch(ctx, opts)
}

func (s *Store) Remove(ctx context.Context, opts plugins.RemoveOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.Store.Remove(ctx, opts)
}

func (s *Store) Update(ctx context.Context, opts plugins.UpdateOptions) error {
	if err := s.Connect(ctx); err != nil {
		return err
	}

	return s.Store.Update(ctx, opts)
}

func EnsureMongoIsRunning(ctx context.Context, c *portercontext.Context, container string, port string, dataVol string, dbName string, timeoutSeconds int) (*mongodb.Store, error) {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	if err := checkDockerAvailability(ctx); err != nil {
		return nil, span.Error(errors.New("Docker is not available"))
	}

	if dataVol != "" {
		err := exec.Command("docker", "volume", "inspect", dataVol).Run()
		if err != nil {
			span.Debugf("creating a data volume, %s, for the mongodb plugin", dataVol)

			err = exec.Command("docker", "volume", "create", dataVol).Run()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					err = fmt.Errorf("%s", string(exitErr.Stderr))
				}
				return nil, span.Error(fmt.Errorf("error creating %s docker volume: %w", dataVol, err))
			}
		}
	}

	mongoImg := "mongo:8.0-noble"

	// TODO(carolynvs): run this using the docker library
	startMongo := func() error {
		span.Debugf("pulling %s", mongoImg)

		err := exec.Command("docker", "pull", mongoImg).Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				err = fmt.Errorf("%s", string(exitErr.Stderr))
			}
			return span.Error(fmt.Errorf("error pulling %s: %w", mongoImg, err))
		}

		span.Debugf("running a mongo server in a container on port %s", port)

		args := []string{"run", "--name", container, "-p=" + port + ":27017", "-d",
			"--health-cmd", "echo 'db.runCommand(\"ping\").ok' | mongosh localhost:27017/admin --quiet",
			"--health-interval", "30s",
			"--health-retries", "3",
			"--health-start-period", "10s",
			"--health-start-interval", "0.5s",
		}
		if dataVol != "" {
			args = append(args, "--mount", "source="+dataVol+",destination=/data/db")
		}
		args = append(args, mongoImg)
		mongoC := exec.Command("docker", args...)
		err = mongoC.Start()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				err = fmt.Errorf("%s", string(exitErr.Stderr))
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
				err = fmt.Errorf("%s", string(exitErr.Stderr))
			}
			return nil, span.Error(fmt.Errorf("error inspecting container %s: %w", container, err))
		}
	} else if !strings.Contains(string(containerStatus), `"Status": "running"`) { // Container is stopped
		err = exec.Command("docker", "rm", "-f", container).Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				err = fmt.Errorf("%s", string(exitErr.Stderr))
			}
			return nil, span.Error(fmt.Errorf("error cleaning up stopped container %s: %w", container, err))
		}

		if err = startMongo(); err != nil {
			return nil, span.Error(err)
		}
	} else if !strings.Contains(string(containerStatus), mongoImg) {
		err = span.Errorf("this version of Porter requires %s. Please upgrade the MongoDB data format as described in https://porter.sh/docs/operations/upgrade-mongo-data-format/.", mongoImg)
		return nil, err
	}

	// wait until the mongo daemon is ready
	span.Debug("waiting for the mongo service to be ready")

	mongoPluginCfg := mongodb.PluginConfig{
		URL:     fmt.Sprintf("mongodb://localhost:%s/%s?connect=direct", port, dbName),
		Timeout: timeoutSeconds,
	}
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	defer cancel()
	for {
		select {
		case <-timeout.Done():
			return nil, span.Error(errors.New("timeout waiting for local mongodb daemon to be ready"))
		case <-tick.C:
			containerStatus, err := exec.Command("docker", "inspect", "--format", "{{lower .State.Health.Status }}", container).Output()
			if err != nil {
				continue
			}
			containerHealth := strings.TrimSpace(string(containerStatus))
			span.Debugf("MongoDB container status: [%s]", containerHealth)
			if strings.EqualFold(containerHealth, "healthy") {
				conn := mongodb.NewStore(c, mongoPluginCfg)
				err = conn.Connect(ctx)
				if err == nil {
					return conn, nil
				}
			} else if strings.EqualFold(containerHealth, "unhealthy") {
				if checkMongoVersionError(container) {
					return nil, span.Errorf("this version of Porter requires %s. Please upgrade the MongoDB data format as described in https://porter.sh/docs/operations/upgrade-mongo-data-format/.", mongoImg)
				}
			} else {
				continue
			}
		}
	}

}

func checkMongoVersionError(container string) bool {
	containerLogs, err := exec.Command("docker", "logs", container).Output()
	if err == nil && strings.Contains(string(containerLogs), "This version of MongoDB is too recent to start up on the existing data files") {
		return true
	}
	return false
}

func checkDockerAvailability(ctx context.Context) error {
	_, err := exec.Command("docker", "info").Output()
	return err
}
