package docker

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
)

const (
	// Container name of the local registry
	DefaultRegistryName = "registry"
)

// Ensure the docker daemon is started and ready to accept connections.
func StartDocker() error {
	switch runtime.GOOS {
	case "windows":
		err := shx.RunS("powershell", "-c", "Get-Process 'Docker Desktop'")
		if err != nil {
			fmt.Println("Starting Docker Desktop")
			cmd := shx.Command(`C:\Program Files\Docker\Docker\Docker Desktop.exe`)
			err := cmd.Cmd.Start()
			if err != nil {
				return errors.Wrapf(err, "could not start Docker Desktop")
			}
		}
	}

	ready, err := isDockerReady()
	if err != nil {
		return err
	}

	if ready {
		return nil
	}

	fmt.Println("Waiting for the docker service to be ready")
	cxt, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	for {
		select {
		case <-cxt.Done():
			return errors.New("a timeout was reached waiting for the docker service to become unavailable")
		default:
			// Wait and check again
			// Writing a dot on a single line so the CI logs show our progress, instead of a bunch of dots at the end
			fmt.Println(".")
			time.Sleep(time.Second)

			if ready, _ := isDockerReady(); ready {
				fmt.Println("Docker service is ready!")
				return nil
			}
		}
	}
}

func isDockerReady() (bool, error) {
	err := shx.RunS("docker", "ps")
	if !sh.CmdRan(err) {
		return false, errors.Wrap(err, "could not run docker")
	}

	return err == nil, nil
}

func NetworkExists(name string) bool {
	err := shx.RunE("docker", "network", "inspect", name)
	return err == nil
}

// Start a Docker registry to use with the tests.
func StartDockerRegistry() error {
	mg.SerialDeps(StartDocker)
	if isContainerRunning(getRegistryName()) {
		return nil
	}

	err := RemoveContainer(getRegistryName())
	if err != nil {
		return err
	}

	fmt.Println("Starting local docker registry")
	return shx.RunE("docker", "run", "-d", "-p", "0.0.0.0:5000:5000", "--name", getRegistryName(), "registry:2")
}

// Stop the Docker registry used by the tests.
func StopDockerRegistry() error {
	if containerExists(getRegistryName()) {
		fmt.Println("Stopping local docker registry")
		return RemoveContainer(getRegistryName())
	}
	return nil
}

func RestartDockerRegistry() error {
	if err := StopDockerRegistry(); err != nil {
		return err
	}
	return StartDockerRegistry()
}

func isContainerRunning(name string) bool {
	out, _ := shx.OutputS("docker", "container", "inspect", "-f", "{{.State.Running}}", name)
	running, _ := strconv.ParseBool(out)
	return running
}

func containerExists(name string) bool {
	err := shx.RunS("docker", "inspect", name)
	return err == nil
}

// Remove the specified container, if it is present.
func RemoveContainer(name string) error {
	stderr := bytes.Buffer{}
	_, _, err := shx.Command("docker", "rm", "-vf", name).Stderr(&stderr).Stdout(nil).Exec()
	// Gracefully handle the container already being gone
	if err != nil && !strings.Contains(stderr.String(), "No such container") {
		return err
	}
	return nil
}

func getRegistryName() string {
	if name, ok := os.LookupEnv("REGISTRY_NAME"); ok {
		return name
	}
	return DefaultRegistryName
}
