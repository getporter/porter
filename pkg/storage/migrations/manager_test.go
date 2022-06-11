package migrations

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_LoadSchema(t *testing.T) {
	t.Run("valid schema", func(t *testing.T) {
		schema := storage.NewSchema()

		c := config.NewTestConfig(t)
		m := NewTestManager(c)
		defer m.Close()

		err := m.store.Update(context.Background(), CollectionConfig, storage.UpdateOptions{Document: schema, Upsert: true})
		require.NoError(t, err, "Save schema failed")

		err = m.loadSchema(context.Background())
		require.NoError(t, err, "LoadSchema failed")
		assert.NotEmpty(t, m.schema, "Schema should be populated with the file's data")
	})

	t.Run("missing schema, empty home", func(t *testing.T) {
		c := config.NewTestConfig(t)
		m := NewTestManager(c)
		defer m.Close()

		err := m.loadSchema(context.Background())
		require.NoError(t, err, "LoadSchema failed")
		assert.NotEmpty(t, m.schema, "Schema should be initialized automatically when PORTER_HOME is empty")
	})

	t.Run("missing schema, existing home data", func(t *testing.T) {
		c := config.NewTestConfig(t)
		m := NewTestManager(c)
		defer m.Close()

		err := m.store.Insert(context.Background(), storage.CollectionInstallations, storage.InsertOptions{Documents: []interface{}{storage.Installation{Name: "abc123"}}})
		require.NoError(t, err)

		err = m.loadSchema(context.Background())
		require.NoError(t, err, "LoadSchema failed")
		assert.Empty(t, m.schema, "Schema should be empty because none was loaded")
	})
}

func TestManager_NoMigrationEmptyHome(t *testing.T) {
	config := config.NewTestConfig(t)
	_, home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.Close()

	mgr := NewTestManager(config)
	defer mgr.Close()
	claimStore := storage.NewInstallationStore(mgr)

	_, err := claimStore.ListInstallations(context.Background(), storage.ListOptions{})
	require.NoError(t, err, "ListInstallations failed")

	credStore := storage.NewCredentialStore(mgr, nil)
	_, err = credStore.ListCredentialSets(context.Background(), storage.ListOptions{})
	require.NoError(t, err, "List credentials failed")

	paramStore := storage.NewParameterStore(mgr, nil)
	_, err = paramStore.ListParameterSets(context.Background(), storage.ListOptions{})
	require.NoError(t, err, "List credentials failed")
}

func TestInstallationStorage_HaltOnMigrationRequired(t *testing.T) {
	t.Parallel()

	tc := config.NewTestConfig(t)
	mgr := NewTestManager(tc)
	defer mgr.Close()
	claimStore := storage.NewInstallationStore(mgr)

	schema := storage.NewSchema()
	schema.Installations = "needs-migration"
	err := mgr.store.Update(context.Background(), CollectionConfig, storage.UpdateOptions{Document: schema, Upsert: true})
	require.NoError(t, err, "Save schema failed")

	checkMigrationError := func(t *testing.T, err error) {
		require.Error(t, err, "Operation should halt because a migration is required")
		assert.Contains(t, err.Error(), "The schema of Porter's data is in an older format than supported by this version of Porter", "The error should be a migration error")

		wantVersionComp := `Porter  uses the following database schema:

storage.Schema{ID:"schema", Installations:"1.0.1", Credentials:"1.0.1", Parameters:"1.0.1"}

Your database schema is:

storage.Schema{ID:"schema", Installations:"needs-migration", Credentials:"1.0.1", Parameters:"1.0.1"}`
		assert.Contains(t, err.Error(), wantVersionComp, "the migration error should contain the current and expected db schema")
	}

	t.Run("list", func(t *testing.T) {
		_, err = claimStore.ListInstallations(context.Background(), storage.ListOptions{})
		checkMigrationError(t, err)
	})

	t.Run("read", func(t *testing.T) {
		_, err = claimStore.GetInstallation(context.Background(), "", "mybun")
		checkMigrationError(t, err)
	})

}

func TestClaimStorage_NoMigrationRequiredForEmptyHome(t *testing.T) {
	t.Parallel()

	config := config.NewTestConfig(t)
	_, home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.Close()

	mgr := NewTestManager(config)
	defer mgr.Close()
	claimStore := storage.NewInstallationStore(mgr)

	names, err := claimStore.ListInstallations(context.Background(), storage.ListOptions{})
	require.NoError(t, err, "ListInstallations failed")
	assert.Empty(t, names, "Expected an empty list of installations since porter home is new")
}

func TestCredentialStorage_HaltOnMigrationRequired(t *testing.T) {
	tc := config.NewTestConfig(t)
	mgr := NewTestManager(tc)
	testSecrets := secrets.NewTestSecretsProvider()
	defer mgr.Close()
	credStore := storage.NewTestCredentialProviderFor(t, mgr, testSecrets)

	schema := storage.NewSchema()
	schema.Credentials = "needs-migration"
	err := mgr.store.Update(context.Background(), CollectionConfig, storage.UpdateOptions{Document: schema, Upsert: true})
	require.NoError(t, err, "Save schema failed")

	checkMigrationError := func(t *testing.T, err error) {
		require.Error(t, err, "Operation should halt because a migration is required")
		assert.Contains(t, err.Error(), "The schema of Porter's data is in an older format than supported by this version of Porter", "The error should be a migration error")

		wantVersionComp := `Porter  uses the following database schema:

storage.Schema{ID:"schema", Installations:"1.0.1", Credentials:"1.0.1", Parameters:"1.0.1"}

Your database schema is:

storage.Schema{ID:"schema", Installations:"1.0.1", Credentials:"needs-migration", Parameters:"1.0.1"}`
		assert.Contains(t, err.Error(), wantVersionComp, "the migration error should contain the current and expected db schema")
	}

	t.Run("list", func(t *testing.T) {
		_, err = credStore.ListCredentialSets(context.Background(), storage.ListOptions{})
		checkMigrationError(t, err)
	})

	t.Run("read", func(t *testing.T) {
		_, err = credStore.GetCredentialSet(context.Background(), "", "mybun")
		checkMigrationError(t, err)
	})
}

func TestCredentialStorage_NoMigrationRequiredForEmptyHome(t *testing.T) {
	config := config.NewTestConfig(t)
	_, home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.Close()

	mgr := NewTestManager(config)
	defer mgr.Close()
	testSecrets := secrets.NewTestSecretsProvider()
	credStore := storage.NewTestCredentialProviderFor(t, mgr, testSecrets)

	names, err := credStore.ListCredentialSets(context.Background(), storage.ListOptions{
		Namespace: "",
		Name:      "",
		Labels:    nil,
		Skip:      0,
		Limit:     0,
	})
	require.NoError(t, err, "List failed")
	assert.Empty(t, names, "Expected an empty list of credentials since porter home is new")
}

func TestParameterStorage_HaltOnMigrationRequired(t *testing.T) {
	tc := config.NewTestConfig(t)
	mgr := NewTestManager(tc)
	defer mgr.Close()
	testSecrets := secrets.NewTestSecretsProvider()
	paramStore := storage.NewTestParameterProviderFor(t, mgr, testSecrets)

	schema := storage.NewSchema()
	schema.Parameters = "needs-migration"
	err := mgr.store.Update(context.Background(), CollectionConfig, storage.UpdateOptions{Document: schema, Upsert: true})
	require.NoError(t, err, "Save schema failed")

	checkMigrationError := func(t *testing.T, err error) {
		require.Error(t, err, "Operation should halt because a migration is required")
		assert.Contains(t, err.Error(), "The schema of Porter's data is in an older format than supported by this version of Porter", "The error should be a migration error")

		wantVersionComp := `Porter  uses the following database schema:

storage.Schema{ID:"schema", Installations:"1.0.1", Credentials:"1.0.1", Parameters:"1.0.1"}

Your database schema is:

storage.Schema{ID:"schema", Installations:"1.0.1", Credentials:"1.0.1", Parameters:"needs-migration"}`
		assert.Contains(t, err.Error(), wantVersionComp, "the migration error should contain the current and expected db schema")
	}

	t.Run("list", func(t *testing.T) {
		_, err = paramStore.ListParameterSets(context.Background(), storage.ListOptions{})
		checkMigrationError(t, err)
	})

	t.Run("read", func(t *testing.T) {
		_, err = paramStore.GetParameterSet(context.Background(), "", "mybun")
		checkMigrationError(t, err)
	})
}

func TestParameterStorage_NoMigrationRequiredForEmptyHome(t *testing.T) {
	config := config.NewTestConfig(t)
	_, home := config.TestContext.UseFilesystem()
	config.SetHomeDir(home)
	defer config.Close()

	mgr := NewTestManager(config)
	defer mgr.Close()
	testSecrets := secrets.NewTestSecretsProvider()
	paramStore := storage.NewTestParameterProviderFor(t, mgr, testSecrets)

	names, err := paramStore.ListParameterSets(context.Background(), storage.ListOptions{})
	require.NoError(t, err, "List failed")
	assert.Empty(t, names, "Expected an empty list of parameters since porter home is new")
}
