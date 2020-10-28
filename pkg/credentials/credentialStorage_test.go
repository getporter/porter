package credentials

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/filesystem"
	"github.com/cnabio/cnab-go/credentials"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialStorage_Validate_GoodSources(t *testing.T) {
	s := CredentialStorage{}
	testCreds := credentials.NewCredentialSet("mycreds",
		valuesource.Strategy{
			Source: valuesource.Source{
				Key:   "env",
				Value: "SOME_ENV",
			},
		},
		valuesource.Strategy{
			Source: valuesource.Source{
				Key:   "value",
				Value: "somevalue",
			},
		})

	err := s.Validate(testCreds)
	require.NoError(t, err, "Validate did not return errors")
}

func TestCredentialStorage_Validate_BadSources(t *testing.T) {
	s := CredentialStorage{}
	testCreds := credentials.NewCredentialSet("mycreds",
		valuesource.Strategy{
			Source: valuesource.Source{
				Key:   "wrongthing",
				Value: "SOME_ENV",
			},
		},
		valuesource.Strategy{
			Source: valuesource.Source{
				Key:   "anotherwrongthing",
				Value: "somevalue",
			},
		},
	)

	err := s.Validate(testCreds)
	require.Error(t, err, "Validate returned errors")
}

func TestCredentialStorage_HaltOnMigrationRequired(t *testing.T) {
	config := config.NewTestConfig(t)
	home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	// Add an unmigrated credential
	credDir := filepath.Join(home, "credentials")
	config.FileSystem.Mkdir(credDir, 0755)
	config.TestContext.AddTestFile(filepath.Join("../storage/testdata/credentials", "mybun.json"), filepath.Join(home, "credentials", "mybun.json"))

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	mgr := storage.NewManager(config.Config, dataStore)
	credStore := credentials.NewCredentialStore(mgr)

	var err error
	t.Run("list", func(t *testing.T) {
		_, err = credStore.List()
		require.Error(t, err, "Operation should halt because a migration is required")
		assert.Contains(t, err.Error(), "The schema of Porter's data is in an older format than supported by this version of Porter")
	})

	t.Run("read", func(t *testing.T) {
		_, err = credStore.Read("mybun")
		require.Error(t, err, "Operation should halt because a migration is required")
		assert.Contains(t, err.Error(), "The schema of Porter's data is in an older format than supported by this version of Porter")
	})
}

func TestCredentialStorage_OperationAllowedWhenNoMigrationDetected(t *testing.T) {
	config := config.NewTestConfig(t)
	home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	// Add migrated credentials data
	config.CopyDirectory(filepath.Join("../storage/testdata", "migrated"), home, false)

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	mgr := storage.NewManager(config.Config, dataStore)
	credStore := credentials.NewCredentialStore(mgr)

	names, err := credStore.List()
	require.NoError(t, err, "List failed")
	assert.NotEmpty(t, names, "Expected credential names to be populated")
}

func TestCredentialStorage_NoMigrationRequiredForEmptyHome(t *testing.T) {
	config := config.NewTestConfig(t)
	home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	dataStore := filesystem.NewStore(*config.Config, hclog.NewNullLogger())
	mgr := storage.NewManager(config.Config, dataStore)
	credStore := credentials.NewCredentialStore(mgr)

	names, err := credStore.List()
	require.NoError(t, err, "List failed")
	assert.Empty(t, names, "Expected an empty list of credentials since porter home is new")
}
