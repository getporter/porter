package migrations

import (
	"context"
	"testing"
	"time"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/encoding"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	testmigrations "get.porter.sh/porter/pkg/storage/migrations/testhelpers"
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

		i := storage.Installation{InstallationSpec: storage.InstallationSpec{
			Name: "abc123",
		}}
		err := m.store.Insert(context.Background(), storage.CollectionInstallations, storage.InsertOptions{Documents: []interface{}{i}})
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

storage.Schema{ID:"schema", Installations:"1.0.2", Credentials:"1.0.1", Parameters:"1.1.0"}

Your database schema is:

storage.Schema{ID:"schema", Installations:"needs-migration", Credentials:"1.0.1", Parameters:"1.1.0"}`
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

storage.Schema{ID:"schema", Installations:"1.0.2", Credentials:"1.0.1", Parameters:"1.1.0"}

Your database schema is:

storage.Schema{ID:"schema", Installations:"1.0.2", Credentials:"needs-migration", Parameters:"1.1.0"}`
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

storage.Schema{ID:"schema", Installations:"1.0.2", Credentials:"1.0.1", Parameters:"1.1.0"}

Your database schema is:

storage.Schema{ID:"schema", Installations:"1.0.2", Credentials:"1.0.1", Parameters:"needs-migration"}`
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

func TestInitCacheEntry_IsValid(t *testing.T) {
	t.Run("fresh entry", func(t *testing.T) {
		entry := initCacheEntry{PorterVersion: pkg.Version, CheckedAt: time.Now()}
		assert.True(t, entry.IsValid())
	})

	t.Run("expired entry", func(t *testing.T) {
		entry := initCacheEntry{PorterVersion: pkg.Version, CheckedAt: time.Now().Add(-2 * initCacheTTL)}
		assert.False(t, entry.IsValid())
	})

	t.Run("version mismatch", func(t *testing.T) {
		entry := initCacheEntry{PorterVersion: "some-other-version", CheckedAt: time.Now()}
		assert.False(t, entry.IsValid())
	})

	t.Run("future timestamp is not trusted", func(t *testing.T) {
		// A CheckedAt in the future (e.g. clock skew, or a hand-edited cache
		// file) must not be treated as valid, since time.Since would be
		// negative and always less than the TTL.
		entry := initCacheEntry{PorterVersion: pkg.Version, CheckedAt: time.Now().Add(time.Hour)}
		assert.False(t, entry.IsValid())
	})
}

func TestManager_Connect_WritesInitCache(t *testing.T) {
	c := config.NewTestConfig(t)
	m := NewTestManager(c)
	defer m.Close()
	ctx := context.Background()

	require.NoError(t, m.Connect(ctx), "Connect failed")

	hash, err := connectionHash(m.Config)
	require.NoError(t, err, "connectionHash failed")

	entry, ok := m.loadInitCacheEntry(ctx, hash)
	require.True(t, ok, "Connect should have written a cache entry")
	assert.Equal(t, pkg.Version, entry.PorterVersion)
	assert.Equal(t, m.schema, entry.Schema)
}

func TestManager_Connect_UsesInitCache(t *testing.T) {
	c := config.NewTestConfig(t)
	m := NewTestManager(c)
	defer m.Close()
	ctx := context.Background()

	require.NoError(t, m.Connect(ctx), "first Connect failed")

	// Make the db schema out-of-date without going through a migration. If a
	// second Connect performs a live check instead of trusting the cache,
	// this will cause it to fail.
	badSchema := storage.NewSchema()
	badSchema.Installations = "needs-migration"
	require.NoError(t, m.store.Update(ctx, CollectionConfig, storage.UpdateOptions{Document: badSchema, Upsert: true}))

	// Simulate a subsequent CLI invocation against the same PORTER_HOME and db.
	m.initialized = false

	err := m.Connect(ctx)
	require.NoError(t, err, "Connect should have trusted the on-disk init cache instead of re-checking the schema")
}

func TestConnectionHash_ChangesWithStorageConfig(t *testing.T) {
	c := config.NewTestConfig(t)
	c.Data.DefaultStorage = "dev"
	c.Data.StoragePlugins = []config.StoragePlugin{
		{PluginConfig: config.PluginConfig{
			Name:         "dev",
			PluginSubKey: "mongodb",
			Config:       map[string]interface{}{"url": "mongodb://localhost:27017"},
		}},
	}

	h1, err := connectionHash(c.Config)
	require.NoError(t, err, "connectionHash failed")

	// Same config should hash the same way.
	h1Again, err := connectionHash(c.Config)
	require.NoError(t, err, "connectionHash failed")
	assert.Equal(t, h1, h1Again, "hash should be stable for an unchanged config")

	// Changing the connection details should change the hash.
	c.Data.StoragePlugins[0].Config = map[string]interface{}{"url": "mongodb://otherhost:27017"}
	h2, err := connectionHash(c.Config)
	require.NoError(t, err, "connectionHash failed")
	assert.NotEqual(t, h1, h2, "hash should change when the storage connection config changes")

	// Pointing at a different named storage entirely should also change the hash.
	c.Data.DefaultStorage = "prod"
	c.Data.StoragePlugins = append(c.Data.StoragePlugins, config.StoragePlugin{
		PluginConfig: config.PluginConfig{
			Name:         "prod",
			PluginSubKey: "mongodb",
			Config:       map[string]interface{}{"url": "mongodb://prodhost:27017"},
		},
	})
	h3, err := connectionHash(c.Config)
	require.NoError(t, err, "connectionHash failed")
	assert.NotEqual(t, h2, h3, "hash should change when the default storage name changes")
}

func TestManager_Connect_InitCacheMissOnConfigChange(t *testing.T) {
	c := config.NewTestConfig(t)
	m := NewTestManager(c)
	defer m.Close()
	ctx := context.Background()

	require.NoError(t, m.Connect(ctx), "first Connect failed")

	// Switch to a different named storage connection. This should compute a
	// different cache key, so the entry written above must not be reused.
	c.Data.DefaultStorage = "alt"
	c.Data.StoragePlugins = []config.StoragePlugin{
		{PluginConfig: config.PluginConfig{
			Name:         "alt",
			PluginSubKey: "mongodb",
			Config:       map[string]interface{}{"url": "mongodb://alt-host:27017"},
		}},
	}

	// Make the db schema out-of-date. If the (now-changed) connection hash
	// still matched the cached entry, Connect would wrongly trust it.
	badSchema := storage.NewSchema()
	badSchema.Installations = "needs-migration"
	require.NoError(t, m.store.Update(ctx, CollectionConfig, storage.UpdateOptions{Document: badSchema, Upsert: true}))

	m.initialized = false

	err := m.Connect(ctx)
	require.Error(t, err, "a changed storage connection should not reuse another connection's cache entry")
	assert.Contains(t, err.Error(), "older format than supported")
}

func TestManager_Connect_InitCacheExpires(t *testing.T) {
	c := config.NewTestConfig(t)
	m := NewTestManager(c)
	defer m.Close()
	ctx := context.Background()

	require.NoError(t, m.Connect(ctx), "first Connect failed")

	hash, err := connectionHash(m.Config)
	require.NoError(t, err, "connectionHash failed")

	path, err := m.initCachePath()
	require.NoError(t, err, "initCachePath failed")

	var cacheFile initCacheFile
	require.NoError(t, encoding.UnmarshalFile(m.FileSystem, path, &cacheFile))
	entry := cacheFile.Storage[hash]
	entry.CheckedAt = time.Now().Add(-2 * initCacheTTL)
	cacheFile.Storage[hash] = entry
	require.NoError(t, encoding.MarshalFile(m.FileSystem, path, cacheFile))

	badSchema := storage.NewSchema()
	badSchema.Installations = "needs-migration"
	require.NoError(t, m.store.Update(ctx, CollectionConfig, storage.UpdateOptions{Document: badSchema, Upsert: true}))

	m.initialized = false

	err = m.Connect(ctx)
	require.Error(t, err, "an expired cache entry should not be trusted")
	assert.Contains(t, err.Error(), "older format than supported")
}

func TestManager_Connect_InitCacheIgnoredOnVersionMismatch(t *testing.T) {
	c := config.NewTestConfig(t)
	m := NewTestManager(c)
	defer m.Close()
	ctx := context.Background()

	require.NoError(t, m.Connect(ctx), "first Connect failed")

	hash, err := connectionHash(m.Config)
	require.NoError(t, err, "connectionHash failed")

	path, err := m.initCachePath()
	require.NoError(t, err, "initCachePath failed")

	var cacheFile initCacheFile
	require.NoError(t, encoding.UnmarshalFile(m.FileSystem, path, &cacheFile))
	entry := cacheFile.Storage[hash]
	entry.PorterVersion = "some-other-version"
	cacheFile.Storage[hash] = entry
	require.NoError(t, encoding.MarshalFile(m.FileSystem, path, cacheFile))

	badSchema := storage.NewSchema()
	badSchema.Installations = "needs-migration"
	require.NoError(t, m.store.Update(ctx, CollectionConfig, storage.UpdateOptions{Document: badSchema, Upsert: true}))

	m.initialized = false

	err = m.Connect(ctx)
	require.Error(t, err, "a cache entry from a different Porter version should not be trusted")
	assert.Contains(t, err.Error(), "older format than supported")
}

func TestManager_Migrate_WritesInitCache(t *testing.T) {
	c := testmigrations.CreateLegacyPorterHome(t)
	defer c.Close()
	oldHome, err := c.GetHomeDir()
	require.NoError(t, err, "could not get the home directory")
	ctx, _, _ := c.SetupIntegrationTest()

	m := NewTestManager(c)
	defer m.Close()

	opts := storage.MigrateOptions{
		OldHome:           oldHome,
		OldStorageAccount: "src",
		NewNamespace:      "myns",
	}

	require.NoError(t, m.Migrate(ctx, opts), "Migrate failed")

	hash, err := connectionHash(m.Config)
	require.NoError(t, err, "connectionHash failed")

	entry, ok := m.loadInitCacheEntry(ctx, hash)
	require.True(t, ok, "Migrate should have written a cache entry")
	assert.Equal(t, pkg.Version, entry.PorterVersion)
	assert.Equal(t, m.schema, entry.Schema)
}
