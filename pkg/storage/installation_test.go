package storage

import (
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallation_String(t *testing.T) {
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
	inst := NewInstallation("myns", "mybuns")

	logs := inst.NewRun("logs")
	logs.Bundle = exampleBundle
	logs.BundleReference = "mybuns:v0.1.1"
	logsResult := logs.NewResult(cnab.StatusSucceeded)
	inst.ApplyResult(logs, logsResult)
	assert.Empty(t, inst.Bundle, "Only actions that modify should update the bundle reference")

	// Try to install and fail
	install1 := inst.NewRun("install")
	install1.Bundle = exampleBundle
	install1.BundleReference = "mybuns:v0.1.1"
	installFailed := install1.NewResult(cnab.StatusFailed)
	inst.ApplyResult(install1, installFailed)
	assert.False(t, inst.IsInstalled())
	assert.Equal(t, inst.Status.BundleVersion, install1.Bundle.Version)
	assert.Equal(t, inst.Status.BundleReference, install1.BundleReference)
	assert.Equal(t, inst.Status.RunID, install1.ID)
	assert.Equal(t, inst.Status.ResultID, installFailed.ID)
	assert.Equal(t, inst.Status.ResultStatus, installFailed.Status)

	// Install and succeed
	install2 := inst.NewRun("install")
	installSucceeded := install2.NewResult(cnab.StatusSucceeded)
	inst.ApplyResult(install2, installSucceeded)
	assert.True(t, inst.IsInstalled())

	// Uninstall and fail
	uninstall1 := inst.NewRun("uninstall")
	uninstallFailed := uninstall1.NewResult(cnab.StatusFailed)
	inst.ApplyResult(uninstall1, uninstallFailed)
	assert.False(t, inst.IsUninstalled())

	// Uninstall and succeed
	uninstall2 := inst.NewRun("uninstall")
	uninstallSucceeded := uninstall2.NewResult(cnab.StatusSucceeded)
	inst.ApplyResult(uninstall2, uninstallSucceeded)
	assert.True(t, inst.IsUninstalled())
}
