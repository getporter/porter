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
	"get.porter.sh/porter/pkg/storage/plugins/protov4"
	"get.porter.sh/porter/pkg/tracing"
)

var _ plugins.StorageBackend = &Store{}

const (
	// startupTimeout is how long to wait for the mongodb container to become ready.
	// Shared CI runners can be slow to start mongod.
	startupTimeout = 60 * time.Second

	// logTailLines is the number of container log lines included in startup errors.
	logTailLines = "20"
)

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

// The methods below implement plugins.StorageBackend. They connect to the
// mongodb container, starting it when necessary, and then use the mongodb
// backend.

func (s *Store) Info(ctx context.Context, req *protov4.InfoRequest) (*protov4.InfoResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).Info(ctx, req)
}

func (s *Store) Migrate(ctx context.Context, req *protov4.MigrateRequest) (*protov4.MigrateResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).Migrate(ctx, req)
}

func (s *Store) Put(ctx context.Context, req *protov4.PutRequest) (*protov4.PutResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).Put(ctx, req)
}

func (s *Store) Get(ctx context.Context, req *protov4.GetRequest) (*protov4.GetResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).Get(ctx, req)
}

func (s *Store) List(ctx context.Context, req *protov4.ListRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).List(ctx, req)
}

func (s *Store) Remove(ctx context.Context, req *protov4.RemoveRequest) (*protov4.RemoveResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).Remove(ctx, req)
}

func (s *Store) RemoveInstallation(ctx context.Context, req *protov4.RemoveInstallationRequest) (*protov4.RemoveInstallationResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).RemoveInstallation(ctx, req)
}

func (s *Store) ListRuns(ctx context.Context, req *protov4.ListRunsRequest) (*protov4.ListRunsResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).ListRuns(ctx, req)
}

func (s *Store) GetLastRun(ctx context.Context, req *protov4.InstallationRef) (*protov4.GetResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).GetLastRun(ctx, req)
}

func (s *Store) GetActiveRuns(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).GetActiveRuns(ctx, req)
}

func (s *Store) ListResults(ctx context.Context, req *protov4.ListResultsRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).ListResults(ctx, req)
}

func (s *Store) InsertOutputs(ctx context.Context, req *protov4.InsertOutputsRequest) (*protov4.InsertOutputsResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).InsertOutputs(ctx, req)
}

func (s *Store) ListOutputs(ctx context.Context, req *protov4.ListOutputsRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).ListOutputs(ctx, req)
}

func (s *Store) GetOutput(ctx context.Context, req *protov4.GetOutputRequest) (*protov4.GetResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).GetOutput(ctx, req)
}

func (s *Store) GetLastOutputs(ctx context.Context, req *protov4.InstallationRef) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).GetLastOutputs(ctx, req)
}

func (s *Store) FindInstallationsReferencing(ctx context.Context, req *protov4.FindReferencingRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).FindInstallationsReferencing(ctx, req)
}

func (s *Store) FindDependencyCandidates(ctx context.Context, req *protov4.FindDependencyCandidatesRequest) (*protov4.ListResponse, error) {
	if err := s.Connect(ctx); err != nil {
		return nil, err
	}
	return mongodb.NewBackend(s.Store).FindDependencyCandidates(ctx, req)
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
			// Workaround for mongod refusing to start on Linux 6.19 up to 7.0.13, https://jira.mongodb.org/browse/SERVER-121912
			// The image sets rseq=0, which makes tcmalloc conflict with the kernel, 1 lets glibc own rseq.
			// Remove once mongod no longer blocks on the kernel, https://jira.mongodb.org/browse/SERVER-125742 and SERVER-136848
			"-e", "GLIBC_TUNABLES=glibc.pthread.rseq=1",
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
	timeout, cancel := context.WithTimeout(ctx, startupTimeout)
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	defer cancel()
	lastHealth := "unknown"
	for {
		select {
		case <-timeout.Done():
			return nil, span.Error(fmt.Errorf("timeout waiting for local mongodb daemon to be ready after %s, last health status: %s\n%s", startupTimeout, lastHealth, containerLogTail(container)))
		case <-tick.C:
			containerStatus, err := exec.Command("docker", "inspect", "--format", "{{lower .State.Health.Status}} {{.State.Running}}", container).Output()
			if err != nil {
				continue
			}
			inspectFields := strings.Fields(string(containerStatus))
			if len(inspectFields) != 2 {
				continue
			}
			containerHealth, running := inspectFields[0], inspectFields[1]
			lastHealth = containerHealth
			span.Debugf("MongoDB container status: [%s]", containerHealth)
			if running == "false" {
				return nil, span.Error(fmt.Errorf("the local mongodb container exited before it was ready\n%s", containerLogTail(container)))
			}
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

// containerLogTail returns the last lines of the container logs for use in error messages.
func containerLogTail(container string) string {
	out, err := exec.Command("docker", "logs", "--tail", logTailLines, container).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("unable to retrieve container logs: %s", err)
	}
	return "container logs:\n" + strings.TrimSpace(string(out))
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
