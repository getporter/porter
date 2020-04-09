package cnabprovider

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/stretchr/testify/require"
)

func TestNewDriver_Docker(t *testing.T) {
	t.Run("vanilla docker", func(t *testing.T) {
		d := NewTestRuntime(t)
		driver, err := d.newDriver(DriverNameDocker, "myclaim", ActionArguments{})

		require.NoError(t, err)
		assert.IsType(t, driver, &docker.Driver{})
	})

	t.Run("docker with host access", func(t *testing.T) {
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
		d := NewTestRuntime(t)
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		_, err := d.newDriver("custom-driver", "myclaim", args)

		assert.EqualError(t, err, "allow-docker-host-access was enabled, but the driver is custom-driver")
	})

	t.Run("docker with host access, missing docker daemon", func(t *testing.T) {
		d := NewTestRuntime(t)
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		_, err := d.newDriver(DriverNameDocker, "myclaim", args)
		assert.EqualError(t, err, "allow-docker-host-access was specified but could not detect a local docker daemon running by checking for /var/run/docker.sock")
	})

	// TODO: add tests that check the Docker configuration options based on the provided required extension config.
	// Requires changes in cnab-go (export of configurationOptions field on Driver struct in docker.go)
}
