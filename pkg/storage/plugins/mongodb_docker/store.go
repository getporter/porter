package mongodb_docker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	portercontext "get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/storage/plugins"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb"
	"github.com/pkg/errors"
)

var _ plugins.StoragePlugin = &Store{}

// Store is a storage plugin for porter suitable for running on machines
// that have not configured proper storage, i.e. a mongo database.
// It runs mongodb in a docker container and stores its data in a docker volume.
type Store struct {
	plugins.StorageProtocol
	context *portercontext.Context

	config PluginConfig
}

func NewStore(cxt *portercontext.Context, cfg PluginConfig) *Store {
	return &Store{
		context: cxt,
		config:  cfg,
	}
}

func (s *Store) Connect() error {
	if s.StorageProtocol != nil {
		return nil
	}

	// Run mongo in a container storing its data in a volume
	container := "porter-mongodb-docker-plugin"
	dataVol := container + "-data"

	conn, err := EnsureMongoIsRunning(s.context, container, s.config.Port, dataVol, s.config.Database, s.config.Timeout)
	if err != nil {
		return err
	}

	s.StorageProtocol = conn
	return nil
}

func (s *Store) Close() error {
	// leave the container running for performance purposes
	//exec.Command("docker", "rm", "-f", "porter-mongodb-docker-plugin")

	// close the connection to the mongodb running in the container
	if conn, ok := s.StorageProtocol.(*mongodb.Store); ok {
		return conn.Close()
	}

	s.StorageProtocol = nil

	return nil
}

func EnsureMongoIsRunning(c *portercontext.Context, container string, port string, dataVol string, dbName string, timeoutSeconds int) (*mongodb.Store, error) {
	if dataVol != "" {
		err := exec.Command("docker", "volume", "inspect", dataVol).Run()
		if err != nil {
			if c.Debug {
				fmt.Fprintf(c.Err, "creating a data volume, %s, for the mongodb plugin\n", dataVol)
			}
			err = exec.Command("docker", "volume", "create", dataVol).Run()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					fmt.Fprintf(c.Err, string(exitErr.Stderr))
				}
				return nil, errors.Wrapf(err, "error creating %s docker volume", dataVol)
			}
		}
	}

	// TODO(carolynvs): run this using the docker library
	startMongo := func() error {
		mongoImg := "mongo:4.0-xenial"
		if c.Debug {
			fmt.Fprintln(c.Err, "pulling", mongoImg)
		}
		err := exec.Command("docker", "pull", mongoImg).Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintf(c.Err, string(exitErr.Stderr))
			}
			return errors.Wrapf(err, "error pulling %s", mongoImg)
		}

		if c.Debug {
			fmt.Fprintln(c.Err, "running a mongo server in a container on port", port)
		}
		args := []string{"run", "--name", container, "-p=" + port + ":27017", "-d"}
		if dataVol != "" {
			args = append(args, "--mount", "source="+dataVol+",destination=/data/db")
		}
		args = append(args, mongoImg)
		mongoC := exec.Command("docker", args...)
		err = mongoC.Start()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintf(c.Err, string(exitErr.Stderr))
			}
			return errors.Wrapf(err, "error running a mongo container for the mongodb-docker plugin")
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
				fmt.Fprintf(c.Err, string(exitErr.Stderr))
			}
			return nil, errors.Wrapf(err, "error inspecting container %s", container)
		}
	} else if !strings.Contains(string(containerStatus), `"Status": "running"`) { // Container is stopped
		err = exec.Command("docker", "rm", "-f", container).Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				fmt.Fprintf(c.Err, string(exitErr.Stderr))
			}
			return nil, errors.Wrapf(err, "error cleaning up stopped container %s", container)
		}

		if err = startMongo(); err != nil {
			return nil, err
		}
	}

	// wait until the mongo daemon is ready
	if c.Debug {
		fmt.Fprintln(c.Err, "waiting for the mongo service to be ready")
	}
	mongoPluginCfg := mongodb.PluginConfig{
		URL:     fmt.Sprintf("mongodb://localhost:%s/%s?connect=direct", port, dbName),
		Timeout: timeoutSeconds,
	}
	timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		select {
		case <-timeout.Done():
			return nil, errors.New("timeout waiting for local mongodb daemon to be ready")
		default:
			conn := mongodb.NewStore(c, mongoPluginCfg)
			err := conn.Connect()
			if err == nil {
				return conn, nil
			} else if c.Debug {
				fmt.Fprintln(c.Err, errors.Wrapf(err, "mongodb isn't ready yet"))
			}
		}
	}
}
