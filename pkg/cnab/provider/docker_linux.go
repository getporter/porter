package cnabprovider

import (
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/pkg/errors"
)

func (r *Runtime) getDockerGroupId() (string, error) {
	resp, err := r.NewCommand("getent", "group", "docker").Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", errors.Errorf("error querying for the docker group id: %s", string(exitErr.Stderr))
		}
		return "", errors.Wrapf(err, "error querying for the docker group id")
	}
	output := strings.TrimSpace(string(resp))
	parts := strings.Split(output, ":")
	if len(parts) < 3 {
		return "", errors.Errorf("could not determine the id of the docker group, unexpected output returned from 'getent group docker': '%s'", output)
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
