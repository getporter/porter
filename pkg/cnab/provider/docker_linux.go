package cnabprovider

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

func (r *Runtime) getDockerGroupId() (string, error) {
	resp, err := r.NewCommand(context.Background(), "getent", "group", "docker").Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("error querying for the docker group id: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("error querying for the docker group id: %w", err)
	}
	output := strings.TrimSpace(string(resp))
	parts := strings.Split(output, ":")
	if len(parts) < 3 {
		return "", fmt.Errorf("could not determine the id of the docker group, unexpected output returned from 'getent group docker': '%s'", output)
	}

	// The command should return GROUP:x:GID
	return parts[2], nil
}

func (r *Runtime) mountDockerSocket(cfg *container.Config, hostCfg *container.HostConfig) error {
	// Add the container to the docker group so that it can access the docker socket
	dockerGID, err := r.getDockerGroupId()
	if err != nil {
		return err
	}
	hostCfg.GroupAdd = []string{dockerGID}

	// Equivalent of using: -v /var/run/docker.sock:/var/run/docker.sock
	// Required for DooD, or "Docker-out-of-Docker"
	dockerSockMount := mount.Mount{
		Source:   "/var/run/docker.sock",
		Target:   "/var/run/docker.sock",
		Type:     "bind",
		ReadOnly: false,
	}
	hostCfg.Mounts = append(hostCfg.Mounts, dockerSockMount)
	return nil
}
