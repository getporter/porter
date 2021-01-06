package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/cnab/extensions"
	"github.com/stretchr/testify/assert"

	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/stretchr/testify/require"
)

func TestNewDriver_Docker(t *testing.T) {
	t.Parallel()

	t.Run("vanilla docker", func(t *testing.T) {
		t.Parallel()

		d := NewTestRuntime(t)
		driver, err := d.newDriver(DriverNameDocker, "myclaim", ActionArguments{})

		require.NoError(t, err)
		assert.IsType(t, driver, &docker.Driver{})
	})

	t.Run("docker with host access", func(t *testing.T) {
		t.Parallel()

		d := NewTestRuntime(t)
		d.FileSystem.Create("/var/run/docker.sock")
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		driver, err := d.newDriver(DriverNameDocker, "myclaim", args)

		require.NoError(t, err)
		assert.IsType(t, driver, &docker.Driver{})
	})

	t.Run("docker with host access, mismatch driver name", func(t *testing.T) {
		t.Parallel()

		d := NewTestRuntime(t)
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		_, err := d.newDriver("custom-driver", "myclaim", args)

		assert.EqualError(t, err, "allow-docker-host-access was enabled, but the driver is custom-driver")
	})

	t.Run("docker with host access, missing docker daemon", func(t *testing.T) {
		t.Parallel()

		d := NewTestRuntime(t)
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		_, err := d.newDriver(DriverNameDocker, "myclaim", args)
		assert.EqualError(t, err, "allow-docker-host-access was specified but could not detect a local docker daemon running by checking for /var/run/docker.sock")
	})

	t.Run("docker with host access, default config", func(t *testing.T) {
		t.Parallel()

		d := NewTestRuntime(t)
		// Currently, toggling Privileged is the only config exposed to users
		// Here we supply no override, so expect Privileged to be false
		d.Extensions[extensions.DockerExtensionKey] = extensions.Docker{}
		d.FileSystem.Create("/var/run/docker.sock")
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		driver, err := d.newDriver(DriverNameDocker, "myclaim", args)
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

		d := NewTestRuntime(t)
		// Currently, toggling Privileged is the only config exposed to users
		// Here we supply an override, so expect Privileged to be set to the override
		d.Extensions[extensions.DockerExtensionKey] = extensions.Docker{
			Privileged: true,
		}
		d.FileSystem.Create("/var/run/docker.sock")
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		driver, err := d.newDriver(DriverNameDocker, "myclaim", args)
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
