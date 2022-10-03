package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDriver_Docker(t *testing.T) {
	t.Parallel()

	t.Run("vanilla docker", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Close()

		driver, err := r.newDriver(DriverNameDocker, ActionArguments{})

		require.NoError(t, err)
		assert.IsType(t, driver, &docker.Driver{})
	})

	t.Run("docker with host access", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		// mock retrieving the docker group id on linux
		r.MockGetDockerGroupId()
		defer r.Close()

		r.FileSystem.Create("/var/run/docker.sock")
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		driver, err := r.newDriver(DriverNameDocker, args)

		require.NoError(t, err)
		assert.IsType(t, driver, &docker.Driver{})
	})

	t.Run("docker with host access, mismatch driver name", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		r.MockGetDockerGroupId()
		defer r.Close()

		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		_, err := r.newDriver("custom-driver", args)

		assert.EqualError(t, err, "allow-docker-host-access was enabled, but the driver is custom-driver")
	})

	t.Run("docker with host access, default config", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		r.MockGetDockerGroupId()
		defer r.Close()

		// Currently, toggling Privileged is the only config exposed to users
		// Here we supply no override, so expect Privileged to be false
		r.Extensions[cnab.DockerExtensionKey] = cnab.Docker{}
		r.FileSystem.Create("/var/run/docker.sock")
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		driver, err := r.newDriver(DriverNameDocker, args)
		require.NoError(t, err)
		assert.IsType(t, driver, &docker.Driver{})

		dockerish, ok := driver.(*docker.Driver)
		assert.True(t, ok)

		err = dockerish.ApplyConfigurationOptions()
		assert.NoError(t, err)

		containerHostCfg, err := dockerish.GetContainerHostConfig()
		require.NoError(t, err)
		require.Equal(t, false, containerHostCfg.Privileged)
	})

	t.Run("docker with host access, privileged true", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		r.MockGetDockerGroupId()
		defer r.Close()

		// Currently, toggling Privileged is the only config exposed to users
		// Here we supply an override, so expect Privileged to be set to the override
		r.Extensions[cnab.DockerExtensionKey] = cnab.Docker{
			Privileged: true,
		}
		r.FileSystem.Create("/var/run/docker.sock")
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		driver, err := r.newDriver(DriverNameDocker, args)
		require.NoError(t, err)
		assert.IsType(t, driver, &docker.Driver{})

		dockerish, ok := driver.(*docker.Driver)
		assert.True(t, ok)

		err = dockerish.ApplyConfigurationOptions()
		assert.NoError(t, err)

		containerHostCfg, err := dockerish.GetContainerHostConfig()
		require.NoError(t, err)
		require.Equal(t, true, containerHostCfg.Privileged)
	})
}
