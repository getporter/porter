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

		_, err := r.FileSystem.Create("/var/run/docker.sock")
		require.NoError(t, err)
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
		_, err := r.FileSystem.Create("/var/run/docker.sock")
		require.NoError(t, err)
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
		_, err := r.FileSystem.Create("/var/run/docker.sock")
		require.NoError(t, err)
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

	t.Run("docker with host access, default config, and host volume mounts", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		r.MockGetDockerGroupId()
		defer r.Close()

		r.Extensions[cnab.DockerExtensionKey] = cnab.Docker{}
		_, err := r.FileSystem.Create("/var/run/docker.sock")
		require.NoError(t, err)
		_, err = r.FileSystem.Create("/sourceFolder")
		require.NoError(t, err)
		_, err = r.FileSystem.Create("/sourceFolder2")
		require.NoError(t, err)
		_, err = r.FileSystem.Create("/sourceFolder3")
		require.NoError(t, err)

		args := ActionArguments{
			AllowDockerHostAccess: true,
			HostVolumeMounts: []HostVolumeMountSpec{
				{
					Source: "/sourceFolder",
					Target: "/targetFolder",
				},
				{
					Source:   "/sourceFolder2",
					Target:   "/targetFolder2",
					ReadOnly: true,
				},
				{
					Source:   "/sourceFolder3",
					Target:   "/targetFolder3",
					ReadOnly: false,
				},
			},
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

		require.Len(t, containerHostCfg.Mounts, 4) //includes the docker socket mount
		assert.Equal(t, "/sourceFolder", containerHostCfg.Mounts[1].Source)
		assert.Equal(t, "/targetFolder", containerHostCfg.Mounts[1].Target)
		assert.Equal(t, false, containerHostCfg.Mounts[1].ReadOnly)
		assert.Equal(t, "/sourceFolder2", containerHostCfg.Mounts[2].Source)
		assert.Equal(t, "/targetFolder2", containerHostCfg.Mounts[2].Target)
		assert.Equal(t, true, containerHostCfg.Mounts[2].ReadOnly)
		assert.Equal(t, "/sourceFolder3", containerHostCfg.Mounts[3].Source)
		assert.Equal(t, "/targetFolder3", containerHostCfg.Mounts[3].Target)
		assert.Equal(t, false, containerHostCfg.Mounts[3].ReadOnly)
	})

	t.Run("host volume mount, docker driver, with multiple mounts", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Close()

		_, err := r.FileSystem.Create("/sourceFolder")
		require.NoError(t, err)
		_, err = r.FileSystem.Create("/sourceFolder2")
		require.NoError(t, err)
		_, err = r.FileSystem.Create("/sourceFolder3")
		require.NoError(t, err)

		args := ActionArguments{
			HostVolumeMounts: []HostVolumeMountSpec{
				{
					Source: "/sourceFolder",
					Target: "/targetFolder",
				},
				{
					Source:   "/sourceFolder2",
					Target:   "/targetFolder2",
					ReadOnly: true,
				},
				{
					Source:   "/sourceFolder3",
					Target:   "/targetFolder3",
					ReadOnly: false,
				},
			},
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

		require.Len(t, containerHostCfg.Mounts, 3)
		assert.Equal(t, "/sourceFolder", containerHostCfg.Mounts[0].Source)
		assert.Equal(t, "/targetFolder", containerHostCfg.Mounts[0].Target)
		assert.Equal(t, false, containerHostCfg.Mounts[0].ReadOnly)
		assert.Equal(t, "/sourceFolder2", containerHostCfg.Mounts[1].Source)
		assert.Equal(t, "/targetFolder2", containerHostCfg.Mounts[1].Target)
		assert.Equal(t, true, containerHostCfg.Mounts[1].ReadOnly)
		assert.Equal(t, "/sourceFolder3", containerHostCfg.Mounts[2].Source)
		assert.Equal(t, "/targetFolder3", containerHostCfg.Mounts[2].Target)
		assert.Equal(t, false, containerHostCfg.Mounts[2].ReadOnly)

	})

	t.Run("host volume mount, docker driver, with single mount", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		defer r.Close()

		_, err := r.FileSystem.Create("/sourceFolder")
		require.NoError(t, err)

		args := ActionArguments{
			HostVolumeMounts: []HostVolumeMountSpec{
				{
					Source: "/sourceFolder",
					Target: "/targetFolder",
				},
			},
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

		require.Len(t, containerHostCfg.Mounts, 1)
		assert.Equal(t, "/sourceFolder", containerHostCfg.Mounts[0].Source)
		assert.Equal(t, "/targetFolder", containerHostCfg.Mounts[0].Target)
		assert.Equal(t, false, containerHostCfg.Mounts[0].ReadOnly)
	})

	t.Run("host volume mount, mismatch driver name", func(t *testing.T) {
		t.Parallel()

		r := NewTestRuntime(t)
		r.MockGetDockerGroupId()
		defer r.Close()

		args := ActionArguments{
			HostVolumeMounts: []HostVolumeMountSpec{
				{
					Source: "/sourceFolder",
					Target: "/targetFolder",
				},
			},
		}

		_, err := r.newDriver("custom-driver", args)

		assert.EqualError(t, err, "mount-host-volume was was used to mount a volume, but the driver is custom-driver")
	})
}
