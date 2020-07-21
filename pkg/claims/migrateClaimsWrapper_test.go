package claims

import (
	"path/filepath"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/filesystem"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateClaimsWrapper_MigrateInstallation(t *testing.T) {
	const installation = "example-exec-outputs"

	testcases := []struct {
		name        string
		fileName    string
		migrateName bool
	}{
		{name: "Has installation name", fileName: "has-installation", migrateName: false},
		{name: "Has claim name", fileName: "has-name", migrateName: true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			config := config.NewTestConfig(t)
			home := config.TestContext.UseFilesystem()
			config.SetHomeDir(home)
			defer config.TestContext.Cleanup()

			claimsDir := filepath.Join(home, "claims")
			config.FileSystem.Mkdir(claimsDir, 0755)
			config.TestContext.AddTestFile(filepath.Join("testdata", tc.fileName+".json"), filepath.Join(claimsDir, tc.fileName+".json"))

			dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
			wrapper := newMigrateClaimsWrapper(config.Context, dataStore)
			claimStore := claim.NewClaimStore(crud.NewBackingStore(wrapper), nil, nil)

			c, err := claimStore.ReadLastClaim(installation)
			require.NoError(t, err, "could not read claim")
			require.NotNil(t, c, "claim should be populated")
			assert.Equal(t, installation, c.Installation, "claim.Installation was not populated")

			assert.Contains(t, config.TestContext.GetError(), "!!! Migrating claims data", "the claim should have been migrated")
			if tc.migrateName {
				assert.Contains(t, config.TestContext.GetError(), "claim.Name to claim.Installation", "the claim should have been migrated from Name -> Installation")
			} else {
				assert.NotContains(t, config.TestContext.GetError(), "claim.Name to claim.Installation", "the claim should NOT be migrated")
			}

			// Read a second time, this time there shouldn't be a migration
			config.TestContext.ClearOutputs()
			_, err = claimStore.ReadLastClaim(installation)
			assert.NotContains(t, config.TestContext.GetError(), "!!! Migrating claims data", "the claim should have been migrated a second time")
		})
	}

	t.Run("no migration", func(t *testing.T) {
		config := config.NewTestConfig(t)
		home := config.TestContext.UseFilesystem()
		config.SetHomeDir(home)
		defer config.TestContext.Cleanup()

		config.CopyDirectory(filepath.Join("testdata", "migrated"), home, false)

		dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
		wrapper := newMigrateClaimsWrapper(config.Context, dataStore)
		claimStore := claim.NewClaimStore(crud.NewBackingStore(wrapper), nil, nil)

		c, err := claimStore.ReadLastClaim(installation)
		require.NoError(t, err, "could not read claim")
		require.NotNil(t, c, "claim should be populated")
		assert.Equal(t, installation, c.Installation, "claim.Installation was not populated")
		assert.NotContains(t, config.TestContext.GetError(), "!!! Migrating claims data", "the claim should have been migrated")
	})
}

func TestMigrateClaimsWrapper_List(t *testing.T) {
	config := config.NewTestConfig(t)
	home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	// Mix up migrated and unmigrated claims
	config.CopyDirectory(filepath.Join("testdata", "migrated"), home, false)
	config.TestContext.AddTestFile(filepath.Join("testdata", "upgraded.json"), filepath.Join(home, "claims", "mybun.json"))
	config.FileSystem.Remove(filepath.Join(home, "schema.json")) // trigger a migration

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	wrapper := newMigrateClaimsWrapper(config.Context, dataStore)
	claimStore := claim.NewClaimStore(crud.NewBackingStore(wrapper), nil, nil)

	names, err := claimStore.ListInstallations()
	sort.Strings(names)
	require.NoError(t, err, "could not list installations")
	assert.Equal(t, []string{"example-exec-outputs", "mybun"}, names, "unexpected list of installation names")
}

func TestMigrateClaimsWrapper_Read(t *testing.T) {
	config := config.NewTestConfig(t)
	home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	claimsDir := filepath.Join(home, "claims")
	config.FileSystem.Mkdir(claimsDir, 0755)
	config.TestContext.AddTestFile("testdata/installed.json", filepath.Join(claimsDir, "installed.json"))
	config.TestContext.AddTestFile("testdata/has-installation.json", filepath.Join(claimsDir, "has-installation.json"))

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	wrapper := newMigrateClaimsWrapper(config.Context, dataStore)
	claimStore := claim.NewClaimStore(crud.NewBackingStore(wrapper), nil, nil)

	// Validate that we can migrate and read in the same operation
	i, err := claimStore.ReadInstallation("mybun")
	require.NoError(t, err, "ReadInstallation failed")
	assert.Equal(t, "mybun", i.Name)
	assert.Equal(t, claim.StatusSucceeded, i.GetLastStatus())
}

func TestMigrateClaimsWrapper_MigrateInstall(t *testing.T) {
	config := config.NewTestConfig(t)
	home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	wrapper := newMigrateClaimsWrapper(config.Context, dataStore)
	claimStore := claim.NewClaimStore(crud.NewBackingStore(wrapper), nil, nil)

	claimsDir := filepath.Join(home, "claims")
	config.FileSystem.Mkdir(claimsDir, 0755)
	config.TestContext.AddTestFile("testdata/installed.json", filepath.Join(claimsDir, "installed.json"))

	err := wrapper.MigrateInstallation("installed")
	require.NoError(t, err, "MigrateInstallation failed")

	exists, _ := config.FileSystem.Exists(filepath.Join(claimsDir, "installed.json"))
	assert.False(t, exists, "the original claim should be removed")

	i, err := claimStore.ReadInstallation("mybun")
	require.NoError(t, err, "ReadInstallation of the migrated claim failed")
	assert.Equal(t, "mybun", i.Name)
	assert.Len(t, i.Claims, 1, "expected 1 claim")

	c, err := i.GetLastClaim()
	require.NoError(t, err)
	assert.Equal(t, claim.ActionInstall, c.Action)
	assert.Equal(t, claim.StatusSucceeded, i.GetLastStatus())
}

func TestMigrateClaimsWrapper_MigrateUpgrade(t *testing.T) {
	config := config.NewTestConfig(t)
	home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	wrapper := newMigrateClaimsWrapper(config.Context, dataStore)
	claimStore := claim.NewClaimStore(crud.NewBackingStore(wrapper), nil, nil)

	claimsDir := filepath.Join(home, "claims")
	config.FileSystem.Mkdir(claimsDir, 0755)
	config.TestContext.AddTestFile("testdata/upgraded.json", filepath.Join(claimsDir, "upgraded.json"))

	err := wrapper.MigrateInstallation("upgraded")
	require.NoError(t, err, "MigrateInstallation failed")

	exists, _ := config.FileSystem.Exists(filepath.Join(claimsDir, "upgraded.json"))
	assert.False(t, exists, "the original claim should be removed")

	i, err := claimStore.ReadInstallation("mybun")
	require.NoError(t, err, "ReadInstallation of the migrated claim failed")
	assert.Equal(t, "mybun", i.Name)
	assert.Len(t, i.Claims, 2, "expected 2 claims")

	c, err := i.GetLastClaim()
	require.NoError(t, err)
	assert.Equal(t, claim.ActionUpgrade, c.Action)
	assert.Equal(t, claim.StatusSucceeded, i.GetLastStatus())

	installClaim := i.Claims[0]
	assert.Equal(t, claim.ActionInstall, installClaim.Action)
	assert.Equal(t, claim.StatusUnknown, installClaim.GetStatus())
}
