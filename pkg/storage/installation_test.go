package storage

import (
	"context"
	"strings"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInstallation(t *testing.T) {
	inst := NewInstallation("dev", "mybuns")

	assert.Equal(t, "mybuns", inst.Name, "Name was not set")
	assert.Equal(t, "dev", inst.Namespace, "Namespace was not set")
	assert.NotEmpty(t, inst.Status.Created, "Created was not set")
	assert.NotEmpty(t, inst.Status.Modified, "Modified was not set")
	assert.Equal(t, inst.Status.Created, inst.Status.Modified, "Created and Modified should have the same timestamp")
	assert.Equal(t, SchemaTypeInstallation, inst.SchemaType, "incorrect SchemaType")
	assert.Equal(t, DefaultInstallationSchemaVersion, inst.SchemaVersion, "incorrect SchemaVersion")
	assert.False(t, inst.Uninstalled, "incorrect Uninstalled")
}

func TestInstallation_String(t *testing.T) {
	t.Parallel()

	i := Installation{InstallationSpec: InstallationSpec{Name: "mybun"}}
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
		{name: "version with buildmeta specified", repo: "ghcr.io/getporter/examples/porter-hello", version: "v0.2.0+1234", wantRef: "ghcr.io/getporter/examples/porter-hello:v0.2.0_1234"},
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

	bun := cnab.ExtendedBundle{}
	t.Run("install failed", func(t *testing.T) {
		// try to install a bundle and fail
		inst := NewInstallation("dev", "mybuns")
		run := inst.NewRun(cnab.ActionInstall, bun)
		result := run.NewResult(cnab.StatusFailed)

		inst.ApplyResult(run, result)

		assert.False(t, inst.IsInstalled(), "a failed install should not mark the installation as installed")
		assert.Empty(t, inst.Status.Installed, "the installed timestamp should not be set")
	})

	t.Run("install succeeded", func(t *testing.T) {
		// install a bundle
		inst := NewInstallation("dev", "mybuns")
		run := inst.NewRun(cnab.ActionInstall, bun)
		result := run.NewResult(cnab.StatusSucceeded)

		inst.ApplyResult(run, result)

		assert.True(t, inst.IsInstalled(), "a failed install should not mark the installation as installed")
		assert.Equal(t, &result.Created, inst.Status.Installed, "the installed timestamp should be set to the result timestamp")
	})

	t.Run("populates InstallationInterfaceHash from the run's bundle outputs", func(t *testing.T) {
		inst := NewInstallation("dev", "mybuns")
		run := inst.NewRun(cnab.ActionInstall, bun)
		run.Bundle = bundle.Bundle{Outputs: map[string]bundle.Output{"connstr": {}}}
		result := run.NewResult(cnab.StatusSucceeded)

		inst.ApplyResult(run, result)

		// A fixed digest, not recomputed via OutputsHash, so this test still
		// catches a regression if both call sites change together. OutputsHash's
		// own properties (stability, sensitivity to the name set, ...) are
		// covered by TestInterfaceCandidate_OutputsHash.
		want := "sha256:3fe2d404b52d449a125c542f4fb9fefc0acea7fb7eebf47c7b41cbb1b41492b7"
		assert.Equal(t, want, inst.Status.InstallationInterfaceHash)
	})

	t.Run("does not set InstallationInterfaceHash for a non-modifying action", func(t *testing.T) {
		inst := NewInstallation("dev", "mybuns")
		run := inst.NewRun("status", bun)
		run.Bundle = bundle.Bundle{
			Outputs: map[string]bundle.Output{"connstr": {}},
			Actions: map[string]bundle.Action{"status": {Modifies: false}},
		}
		result := run.NewResult(cnab.StatusSucceeded)

		inst.ApplyResult(run, result)

		assert.Empty(t, inst.Status.InstallationInterfaceHash)
	})

	t.Run("uninstall failed", func(t *testing.T) {
		// Make an installed bundle
		inst := NewInstallation("dev", "mybuns")
		inst.Status.Created = now.Add(-time.Second * 10)
		inst.Status.Installed = &inst.Status.Created

		// try to uninstall it and fail
		run := inst.NewRun(cnab.ActionUninstall, bun)
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
		run := inst.NewRun(cnab.ActionUninstall, bun)
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
		run := inst.NewRun(cnab.ActionUninstall, bun)
		result := run.NewResult(cnab.StatusSucceeded)
		result.Created = now.Add(-time.Second * 10)

		inst.ApplyResult(run, result)

		assert.False(t, inst.IsInstalled(), "the installation should no longer be considered installed")
		assert.True(t, inst.IsUninstalled(), "the installation should be marked as uninstalled")
		assert.Equal(t, &inst.Status.Created, inst.Status.Installed, "the installed timestamp should still be set")
		assert.Equal(t, &result.Created, inst.Status.Uninstalled, "the uninstalled timestamp should be set")

		// re-install the bundle
		run = inst.NewRun(cnab.ActionInstall, bun)
		result = run.NewResult(cnab.StatusSucceeded)
		result.Created = now.Add(-time.Second * 5)

		inst.ApplyResult(run, result)

		assert.True(t, inst.IsInstalled(), "the installation should be marked as installed")
		assert.False(t, inst.IsUninstalled(), "the installation should not be marked as uninstalled")
		assert.Equal(t, &result.Created, inst.Status.Installed, "the installed timestamp should be set to the new install time")
		assert.NotEmpty(t, inst.Status.Uninstalled, "the uninstalled timestamp should still be be set")

		// re-uninstall the bundle
		run = inst.NewRun(cnab.ActionUninstall, bun)
		result = run.NewResult(cnab.StatusSucceeded)

		inst.ApplyResult(run, result)

		assert.False(t, inst.IsInstalled(), "the installation should not be marked as installed")
		assert.True(t, inst.IsUninstalled(), "the installation should be marked as uninstalled")
		assert.NotEmpty(t, inst.Status.Installed, "the installed timestamp should still be be set")
		assert.Equal(t, &result.Created, inst.Status.Uninstalled, "the uninstalled timestamp should be set to the new uninstall time")
	})
}

func TestInstallation_References(t *testing.T) {
	t.Parallel()

	t.Run("AddReference adds a new reference", func(t *testing.T) {
		inst := NewInstallation("dev", "mysqldb")

		changed := inst.AddReference("dev/myinfra", "db")

		assert.True(t, changed, "expected AddReference to report a change")
		assert.Equal(t, []InstallationReference{{Installation: "dev/myinfra", Dependency: "db"}}, inst.Status.References)
		assert.True(t, inst.IsReferenced())
	})

	t.Run("AddReference is idempotent", func(t *testing.T) {
		inst := NewInstallation("dev", "mysqldb")
		inst.AddReference("dev/myinfra", "db")

		changed := inst.AddReference("dev/myinfra", "db")

		assert.False(t, changed, "expected a duplicate AddReference to report no change")
		assert.Len(t, inst.Status.References, 1, "expected the duplicate reference to not be added again")
	})

	t.Run("AddReference tracks multiple referencing installations", func(t *testing.T) {
		inst := NewInstallation("dev", "mysqldb")
		inst.AddReference("dev/myinfra", "db")

		changed := inst.AddReference("dev/webapp", "db")

		assert.True(t, changed)
		assert.Len(t, inst.Status.References, 2)
	})

	t.Run("RemoveReference removes all aliases for an installation", func(t *testing.T) {
		inst := NewInstallation("dev", "mysqldb")
		inst.AddReference("dev/myinfra", "db")
		inst.AddReference("dev/myinfra", "cache")
		inst.AddReference("dev/webapp", "db")

		changed := inst.RemoveReference("dev/myinfra")

		assert.True(t, changed)
		assert.Equal(t, []InstallationReference{{Installation: "dev/webapp", Dependency: "db"}}, inst.Status.References)
	})

	t.Run("RemoveReference is a no-op when not referenced", func(t *testing.T) {
		inst := NewInstallation("dev", "mysqldb")

		changed := inst.RemoveReference("dev/myinfra")

		assert.False(t, changed)
		assert.False(t, inst.IsReferenced())
	})

	t.Run("IsReferenced reflects the current reference list", func(t *testing.T) {
		inst := NewInstallation("dev", "mysqldb")
		assert.False(t, inst.IsReferenced())

		inst.AddReference("dev/myinfra", "db")
		assert.True(t, inst.IsReferenced())

		inst.RemoveReference("dev/myinfra")
		assert.False(t, inst.IsReferenced())
	})
}

func TestInstallation_Validate(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name      string
		input     InstallationSpec
		wantError string
	}{
		{
			name: "none",
			input: InstallationSpec{
				SchemaType:    "",
				SchemaVersion: DefaultInstallationSchemaVersion},
			wantError: ""},
		{
			name: strings.ToLower(SchemaTypeInstallation),
			input: InstallationSpec{
				SchemaType:    "installation",
				SchemaVersion: DefaultInstallationSchemaVersion},
			wantError: ""},
		{
			name: SchemaTypeInstallation,
			input: InstallationSpec{
				SchemaType:    SchemaTypeInstallation,
				SchemaVersion: DefaultInstallationSchemaVersion},
			wantError: ""},
		{
			name: strings.ToUpper(SchemaTypeInstallation),
			input: InstallationSpec{
				SchemaType:    "INSTALLATION",
				SchemaVersion: DefaultInstallationSchemaVersion},
			wantError: ""},
		{
			name: SchemaTypeCredentialSet,
			input: InstallationSpec{
				SchemaType:    SchemaTypeCredentialSet,
				SchemaVersion: DefaultInstallationSchemaVersion},
			wantError: "invalid schemaType CredentialSet, expected Installation"},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			err := tc.input.Validate(ctx, schema.CheckStrategyExact)
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				tests.RequireErrorContains(t, err, tc.wantError)
			}
		})
	}
}

func TestInstallation_Validate_DefaultSchemaType(t *testing.T) {
	i := NewInstallation("", "mybuns")
	i.SchemaType = ""
	require.NoError(t, i.Validate(context.Background(), schema.CheckStrategyExact))
	assert.Equal(t, SchemaTypeInstallation, i.SchemaType)
}
