package migrations

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage/migrations/crudstore"
	testmigrations "get.porter.sh/porter/pkg/storage/migrations/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verify that we can retrieve data from the old plugins
func TestLegacyPluginAdapter(t *testing.T) {
	c := testmigrations.CreateLegacyPorterHome(t)
	defer c.Close()

	home, err := c.GetHomeDir()
	require.NoError(t, err, "could not get the home directory")

	adapter := NewLegacyPluginAdapter(c.Config, home, "src")
	defer adapter.Close()

	ctx := context.Background()

	// List installation names
	installationNames, err := adapter.List(ctx, "installations", "")
	require.NoError(t, err, "failed to list the installation names")
	wantInstallationNames := []string{"creds-tutorial", "hello-llama", "hello1", "sensitive-data"}
	require.Equal(t, wantInstallationNames, installationNames, "expected 4 installations")

	// Retrieve a claim document
	result, err := adapter.Read(ctx, "claims", "01G1VJQJV0RN5AW5BSZHNTVYTV")
	require.NoError(t, err, "failed to read the claim document")

	// Check that we did read the claim correctly through the plugin
	var gotData map[string]interface{}
	err = json.Unmarshal(result, &gotData)
	require.NoError(t, err, "failed to unmarshal the claim document")

	var wantData map[string]interface{}
	contents, err := ioutil.ReadFile(filepath.Join(home, "claims/hello1/01G1VJQJV0RN5AW5BSZHNTVYTV.json"))
	require.NoError(t, err, "error reading the test claim to compare results")
	require.NoError(t, json.Unmarshal(contents, &wantData), "failed to unmarshal the test claim")
	assert.Equal(t, wantData, gotData, "The claim data read through the plugin doesn't match the original test claim")
}

func TestLegacyPluginAdapter_makePluginConfig(t *testing.T) {
	c := testmigrations.CreateLegacyPorterHome(t)
	defer c.Close()

	home, err := c.GetHomeDir()
	require.NoError(t, err, "could not get the home directory")

	adapter := NewLegacyPluginAdapter(c.Config, home, "src")
	defer adapter.Close()

	cfg := adapter.makePluginConfig()

	// Test that we are using the legacy plugin interface
	assert.Equal(t, "storage", cfg.Interface, "incorrect plugin interface")
	assert.Equal(t, &crudstore.Plugin{}, cfg.Plugin, "incorrect plugin wrapper")
	assert.Equal(t, uint(1), cfg.ProtocolVersion, "incorrect plugin protocol version")

	// Test that we are using the correct default plugin based on porter v0
	// Use an empty config since we are testing defaults
	defaultPlugin := cfg.GetDefaultPlugin(&config.Config{})
	assert.Equal(t, "filesystem", defaultPlugin, "incorrect default storage plugin")

	// Test that we are using the storage account specified by the user
	// Use an empty config since we are testing defaults
	defaultStorageName := cfg.GetDefaultPluggable(&config.Config{})
	assert.Equal(t, "src", defaultStorageName, "incorrect default storage account used")

	// Test that we can retrieve a named legacy storage account
	// Use the real config from our temp PORTER_HOME, which has the azure storage account defined
	plugin, err := cfg.GetPluggable(c.Config, "azure")
	require.NoError(t, err, "GetPluggable failed")
	assert.Equal(t, "azure", plugin.GetName(), "incorrect plugin retrieved")
}
