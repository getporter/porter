package cnabprovider

import (
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

func (r *Runtime) mountDockerSocket(cfg *container.Config, hostCfg *container.HostConfig) error {
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

func (r *Runtime) addVolumeMountToHostConfig(hostConfig *container.HostConfig, source string, target string, readOnly bool) error {
	mount := mount.Mount{
		Source:   source,
		Target:   target,
		Type:     "bind",
		ReadOnly: readOnly,
	}
	hostConfig.Mounts = append(hostConfig.Mounts, mount)
	return nil
}
