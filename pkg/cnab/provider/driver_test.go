package cnabprovider

import (
	"os"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/cnab/extensions"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"

	"github.com/cnabio/cnab-go/driver/docker"
	"github.com/stretchr/testify/require"
)

func TestNewDriver_Docker(t *testing.T) {
	// Do not run in parallel as it manipulates environment variables
	assertPullAlways := func(t *testing.T, shouldPull bool) {
		if shouldPull {
			assert.Equal(t, "1", os.Getenv("PULL_ALWAYS"),  "PULL_ALWAYS should be set")
		} else {
			assert.NotEqual(t, "1", os.Getenv("PULL_ALWAYS"),  "PULL_ALWAYS should NOT be set")
		}
	}

	t.Run("do not pull if missing repo digests", func(t *testing.T) {
		os.Unsetenv("PULL_ALWAYS")

		r := NewTestRuntime(t)
		defer r.Teardown()

		b := bundle.Bundle{InvocationImages: []bundle.InvocationImage{
			{BaseImage: bundle.BaseImage{
				Image: "getporter/porter-hello-installer",
				Digest: "",
			}},
		}}
		args := ActionArguments{
			BundleReference: cnab.BundleReference{Definition: b},
		}
		driver, err := r.newDriver(DriverNameDocker, args)

		require.NoError(t, err)
		require.IsType(t, driver, &docker.Driver{})
		assertPullAlways(t, false)
	})

	t.Run("pull image when repo digest set", func(t *testing.T) {
		os.Unsetenv("PULL_ALWAYS")

		r := NewTestRuntime(t)
		defer r.Teardown()

		// Pull the image because we have a digest set
		b := bundle.Bundle{InvocationImages: []bundle.InvocationImage{
			{BaseImage: bundle.BaseImage{
				Image: "getporter/porter-hello-installer",
				Digest: "sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687",
			}},
		}}
		args := ActionArguments{
			BundleReference: cnab.BundleReference{Definition: b},
		}
		driver, err := r.newDriver(DriverNameDocker, args)

		require.NoError(t, err)
		require.IsType(t, driver, &docker.Driver{})
		assertPullAlways(t, true)
	})

	t.Run("vanilla docker", func(t *testing.T) {
		r := NewTestRuntime(t)
		defer r.Teardown()

		driver, err := r.newDriver(DriverNameDocker, ActionArguments{})

		require.NoError(t, err)
		require.IsType(t, driver, &docker.Driver{})
	})

	t.Run("docker with host access", func(t *testing.T) {
		r := NewTestRuntime(t)
		defer r.Teardown()

		r.FileSystem.Create("/var/run/docker.sock")
		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		driver, err := r.newDriver(DriverNameDocker, args)

		require.NoError(t, err)
		assert.IsType(t, driver, &docker.Driver{})
	})

	t.Run("docker with host access, mismatch driver name", func(t *testing.T) {
		r := NewTestRuntime(t)
		defer r.Teardown()

		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		_, err := r.newDriver("custom-driver", args)

		assert.EqualError(t, err, "allow-docker-host-access was enabled, but the driver is custom-driver")
	})

	t.Run("docker with host access, missing docker daemon", func(t *testing.T) {
		r := NewTestRuntime(t)
		defer r.Teardown()

		args := ActionArguments{
			AllowDockerHostAccess: true,
		}

		_, err := r.newDriver(DriverNameDocker, args)
		assert.EqualError(t, err, "allow-docker-host-access was specified but could not detect a local docker daemon running by checking for /var/run/docker.sock")
	})

	t.Run("docker with host access, default config", func(t *testing.T) {
		r := NewTestRuntime(t)
		defer r.Teardown()

		// Currently, toggling Privileged is the only config exposed to users
		// Here we supply no override, so expect Privileged to be false
		r.Extensions[extensions.DockerExtensionKey] = extensions.Docker{}
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
		r := NewTestRuntime(t)
		defer r.Teardown()

		// Currently, toggling Privileged is the only config exposed to users
		// Here we supply an override, so expect Privileged to be set to the override
		r.Extensions[extensions.DockerExtensionKey] = extensions.Docker{
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
