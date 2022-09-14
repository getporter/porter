package storage

import (
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallation_String(t *testing.T) {
	t.Parallel()

	i := Installation{Name: "mybun"}
	assert.Equal(t, "/mybun", i.String())

	i.Namespace = "dev"
	assert.Equal(t, "dev/mybun", i.String())
}

func TestOCIReferenceParts_GetBundleReference(t *testing.T) {
	testcases := []struct {
		name    string
		repo    string
		digest  string
		version string
		tag     string
		wantRef string
		wantErr string
	}{
		{name: "repo missing", wantRef: ""},
		{name: "incomplete reference", repo: "ghcr.io/getporter/examples/porter-hello", wantErr: "Invalid bundle reference"},
		{name: "version specified", repo: "ghcr.io/getporter/examples/porter-hello", version: "v0.2.0", wantRef: "ghcr.io/getporter/examples/porter-hello:v0.2.0"},
		{name: "digest specified", repo: "ghcr.io/getporter/examples/porter-hello", digest: "sha256:a881bbc015bade9f11d95a4244888d8e7fa8800f843b43c74cc07c7b7276b062", wantRef: "ghcr.io/getporter/examples/porter-hello@sha256:a881bbc015bade9f11d95a4244888d8e7fa8800f843b43c74cc07c7b7276b062"},
		{name: "tag specified", repo: "ghcr.io/getporter/examples/porter-hello", tag: "latest", wantRef: "ghcr.io/getporter/examples/porter-hello:latest"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b := OCIReferenceParts{
				Repository: tc.repo,
				Digest:     tc.digest,
				Version:    tc.version,
				Tag:        tc.tag,
			}

			ref, ok, err := b.GetBundleReference()
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			} else if tc.wantRef != "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantRef, ref.String())
			} else {
				require.NoError(t, err)
				require.False(t, ok)
			}
		})
	}
}

func TestInstallation_ApplyResult(t *testing.T) {
	t.Parallel()

	t.Run("install failed", func(t *testing.T) {
		// try to install a bundle and fail
		inst := NewInstallation("dev", "mybuns")
		run := inst.NewRun(cnab.ActionInstall)
		result := run.NewResult(cnab.StatusFailed)

		inst.ApplyResult(run, result)

		assert.False(t, inst.IsInstalled(), "a failed install should not mark the installation as installed")
		assert.Empty(t, inst.Status.Installed, "the installed timestamp should not be set")
	})

	t.Run("install succeeded", func(t *testing.T) {
		// install a bundle
		inst := NewInstallation("dev", "mybuns")
		run := inst.NewRun(cnab.ActionInstall)
		result := run.NewResult(cnab.StatusSucceeded)

		inst.ApplyResult(run, result)

		assert.True(t, inst.IsInstalled(), "a failed install should not mark the installation as installed")
		assert.Equal(t, &result.Created, inst.Status.Installed, "the installed timestamp should be set to the result timestamp")
	})

	t.Run("uninstall failed", func(t *testing.T) {
		// Make an installed bundle
		inst := NewInstallation("dev", "mybuns")
		inst.Status.Created = now.Add(-time.Second * 10)
		inst.Status.Installed = &inst.Status.Created

		// try to uninstall it and fail
		run := inst.NewRun(cnab.ActionUninstall)
		result := run.NewResult(cnab.StatusFailed)

		inst.ApplyResult(run, result)

		assert.True(t, inst.IsInstalled(), "the installation should still be marked as installed")
		assert.False(t, inst.IsUninstalled(), "the installation should not be marked as uninstalled")
		assert.Empty(t, inst.Status.Uninstalled, "the uninstalled timestamp should not be set")
	})

	t.Run("uninstall succeeded", func(t *testing.T) {
		// Make an installed bundle
		inst := NewInstallation("dev", "mybuns")
		inst.Status.Created = now.Add(-time.Second * 10)
		inst.Status.Installed = &inst.Status.Created

		// uninstall it
		run := inst.NewRun(cnab.ActionUninstall)
		result := run.NewResult(cnab.StatusSucceeded)

		inst.ApplyResult(run, result)

		assert.False(t, inst.IsInstalled(), "the installation should no longer be considered installed")
		assert.True(t, inst.IsUninstalled(), "the installation should be marked as uninstalled")
		assert.Equal(t, &inst.Status.Created, inst.Status.Installed, "the installed timestamp should still be set")
		assert.Equal(t, &result.Created, inst.Status.Uninstalled, "the uninstalled timestamp should be set")
	})

	t.Run("desired state after re-installation and re-unstallation", func(t *testing.T) {
		// Make an installed bundle
		inst := NewInstallation("dev", "mybuns")
		inst.Status.Created = now.Add(-time.Second * 15)
		inst.Status.Installed = &inst.Status.Created

		// uninstall the bundle
		run := inst.NewRun(cnab.ActionUninstall)
		result := run.NewResult(cnab.StatusSucceeded)
		result.Created = now.Add(-time.Second * 10)

		inst.ApplyResult(run, result)

		assert.False(t, inst.IsInstalled(), "the installation should no longer be considered installed")
		assert.True(t, inst.IsUninstalled(), "the installation should be marked as uninstalled")
		assert.Equal(t, &inst.Status.Created, inst.Status.Installed, "the installed timestamp should still be set")
		assert.Equal(t, &result.Created, inst.Status.Uninstalled, "the uninstalled timestamp should be set")

		// re-install the bundle
		run = inst.NewRun(cnab.ActionInstall)
		result = run.NewResult(cnab.StatusSucceeded)
		result.Created = now.Add(-time.Second * 5)

		inst.ApplyResult(run, result)

		assert.True(t, inst.IsInstalled(), "the installation should be marked as installed")
		assert.False(t, inst.IsUninstalled(), "the installation should not be marked as uninstalled")
		assert.Equal(t, &result.Created, inst.Status.Installed, "the installed timestamp should be set to the new install time")
		assert.NotEmpty(t, inst.Status.Uninstalled, "the uninstalled timestamp should still be be set")

		// re-uninstall the bundle
		run = inst.NewRun(cnab.ActionUninstall)
		result = run.NewResult(cnab.StatusSucceeded)

		inst.ApplyResult(run, result)

		assert.False(t, inst.IsInstalled(), "the installation should not be marked as installed")
		assert.True(t, inst.IsUninstalled(), "the installation should be marked as uninstalled")
		assert.NotEmpty(t, inst.Status.Installed, "the installed timestamp should still be be set")
		assert.Equal(t, &result.Created, inst.Status.Uninstalled, "the uninstalled timestamp should be set to the new uninstall time")
	})
}
