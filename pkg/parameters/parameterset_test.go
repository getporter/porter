package parameters

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/filesystem"
	"github.com/cnabio/cnab-go/schema"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCNABSpecVersion(t *testing.T) {
	version, err := schema.GetSemver(CNABSpecVersion)
	require.NoError(t, err)
	assert.Equal(t, DefaultSchemaVersion, version)
}

func TestNewParameterSet(t *testing.T) {
	cs := NewParameterSet("myparams",
		valuesource.Strategy{
			Name: "password",
			Source: valuesource.Source{
				Key:   "env",
				Value: "DB_PASSWORD",
			},
		})

	assert.Equal(t, "myparams", cs.Name, "Name was not set")
	assert.NotEmpty(t, cs.Created, "Created was not set")
	assert.NotEmpty(t, cs.Modified, "Modified was not set")
	assert.Equal(t, cs.Created, cs.Modified, "Created and Modified should have the same timestamp")
	assert.Equal(t, DefaultSchemaVersion, cs.SchemaVersion, "SchemaVersion was not set")
	assert.Len(t, cs.Parameters, 1, "Parameters should be initialized with 1 value")
}

// TODO: (carolynvs) move this into manager_test.go in pkg/storage once parameter set is moved to cnab-go
func TestManager_MigrateParameters(t *testing.T) {
	config := config.NewTestConfig(t)
	home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.TestContext.Cleanup()

	credsDir := filepath.Join(home, "parameters")
	config.FileSystem.Mkdir(credsDir, 0755)
	config.TestContext.AddTestFile(filepath.Join("../storage/testdata/parameters", "mybun.json"), filepath.Join(credsDir, "mybun.json"))

	dataStore := crud.NewBackingStore(filesystem.NewStore(*config.Config, hclog.NewNullLogger()))
	mgr := storage.NewManager(config.Config, dataStore)
	paramStore := NewParameterStorage(mgr)

	logfilePath, err := mgr.Migrate()
	require.NoError(t, err, "Migrate failed")

	c, err := paramStore.Read("mybun")
	require.NoError(t, err, "Read parameter failed")

	assert.Equal(t, DefaultSchemaVersion, c.SchemaVersion, "parameter set was not migrated")

	logfile, err := config.FileSystem.ReadFile(logfilePath)
	require.NoError(t, err, "error reading logfile")
	assert.Equal(t, config.TestContext.GetError(), string(logfile), "the migration should have been copied to both stderr and the logfile")
}
