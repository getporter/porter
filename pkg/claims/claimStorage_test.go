package claims

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/filesystem"
	"github.com/cnabio/cnab-go/claim"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClaimStorage_HaltOnMigrationRequired(t *testing.T) {
	config := config.NewTestConfig(t)
	_, home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	// Add an unmigrated claim
	claimsDir := filepath.Join(home, "claims")
	config.FileSystem.Mkdir(claimsDir, 0755)
	config.TestContext.AddTestFile(filepath.Join("../storage/testdata/claims", "upgraded.json"), filepath.Join(home, "claims", "mybun.json"))

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	mgr := storage.NewManager(config.Config, dataStore)
	claimStore := claim.NewClaimStore(mgr, nil, nil)

	var err error
	t.Run("list", func(t *testing.T) {
		_, err = claimStore.ListInstallations()
		require.Error(t, err, "Operation should halt because a migration is required")
		assert.Contains(t, err.Error(), "The schema of Porter's data is in an older format than supported by this version of Porter")
	})

	t.Run("read", func(t *testing.T) {
		_, err = claimStore.ReadInstallation("mybun")
		require.Error(t, err, "Operation should halt because a migration is required")
		assert.Contains(t, err.Error(), "The schema of Porter's data is in an older format than supported by this version of Porter")
	})

}

func TestClaimStorage_OperationAllowedWhenNoMigrationDetected(t *testing.T) {
	config := config.NewTestConfig(t)
	_, home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	// Add migrated claims data
	config.CopyDirectory(filepath.Join("../storage/testdata", "migrated"), home, false)

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	mgr := storage.NewManager(config.Config, dataStore)
	claimStore := claim.NewClaimStore(mgr, nil, nil)

	names, err := claimStore.ListInstallations()
	require.NoError(t, err, "ListInstallations failed")
	assert.NotEmpty(t, names, "Expected installation names to be populated")
}

func TestClaimStorage_NoMigrationRequiredForEmptyHome(t *testing.T) {
	config := config.NewTestConfig(t)
	_, home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	mgr := storage.NewManager(config.Config, dataStore)
	claimStore := claim.NewClaimStore(mgr, nil, nil)

	names, err := claimStore.ListInstallations()
	require.NoError(t, err, "ListInstallations failed")
	assert.Empty(t, names, "Expected an empty list of installations since porter home is new")
}
