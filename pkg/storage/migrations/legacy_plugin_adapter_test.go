package migrations

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/carolynvs/magex/pkg/downloads"
	"github.com/carolynvs/magex/shx"
	"github.com/carolynvs/magex/xplat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Verify that we can retrieve data from the old plugins
func TestLegacyPluginAdapter(t *testing.T) {
	c := createLegacyPorterHome(t)
	defer c.Close()

	home, err := c.GetHomeDir()
	require.NoError(t, err, "could not get the home directory")

	adapter := NewLegacyPluginAdapter(c.Config, home, "src")
	defer adapter.Close()

	ctx := context.Background()

	// List installation names
	installationNames, err := adapter.List(ctx, "installations", "")
	require.NoError(t, err, "failed to list the installation names")
	require.Len(t, installationNames, 1, "expected only one installation name")
	assert.Equal(t, "hello1", installationNames[0], "incorrect installation name")

	// Retrieve a claim document
	result, err := adapter.Read(ctx, "claims", "01G1VJGY43HT3KZN82DS6DDPWK")
	require.NoError(t, err, "failed to read the claim document")

	// Check that we did read the claim correctly through the plugin
	var gotData map[string]interface{}
	err = json.Unmarshal(result, &gotData)
	require.NoError(t, err, "failed to unmarshal the claim document")

	var wantData map[string]interface{}
	contents, err := ioutil.ReadFile("testdata/v0_home/claims/hello1/01G1VJGY43HT3KZN82DS6DDPWK.json")
	require.NoError(t, err, "error reading the test claim to compare results")
	require.NoError(t, json.Unmarshal(contents, &wantData), "failed to unmarshal the test claim")
	assert.Equal(t, wantData, gotData, "The claim data read through the plugin doesn't match the original test claim")
}

// create a porter v0.38 PORTER_HOME with existing filesystem storage plugin data
func createLegacyPorterHome(t *testing.T) *config.TestConfig {
	tmp, err := ioutil.TempDir("", "porter")
	require.NoError(t, err)

	c := config.NewTestConfig(t)
	c.TestContext.UseFilesystem()
	c.SetHomeDir(tmp)
	c.TestContext.AddCleanupDir(tmp) // Remove the temp home directory when the context is closed
	err = c.CopyDirectory("testdata/v0_home", tmp, false)
	require.NoError(t, err, "error copying testdata into the temporary PORTER_HOME")

	// Download a v0.38 release into our old porter home
	// Cache it in bin/v0 so that we don't re-download it unnecessarily
	oldPorterDir := filepath.Join(c.TestContext.FindRepoRoot(), "bin", "v0")
	oldPorterPath := filepath.Join(oldPorterDir, "porter"+xplat.FileExt())
	if _, err = os.Stat(oldPorterPath); os.IsNotExist(err) {
		os.MkdirAll(oldPorterDir, 0700)
		err = downloads.Download(oldPorterDir, downloads.DownloadOptions{
			UrlTemplate: "https://github.com/getporter/porter/releases/download/{{.VERSION}}/porter-{{.GOOS}}-amd64",
			Name:        "porter",
			Version:     "v0.38.10",
		})
		require.NoError(t, err, "Failed to download a copy of the old version of porter")
	}

	// Put the old version of porter into PORTER_HOME
	err = shx.Copy(filepath.Join(oldPorterDir, "porter"+xplat.FileExt()), tmp)
	require.NoError(t, err, "Failed to copy the old version of porter into the temp PORTER_HOME")

	// fixup permissions on the home directory to make the filesystem plugin happy
	// I'm calling porter so that I don't reimplement this functionality
	shx.Command(oldPorterPath, "storage", "fix-permissions").
		Env("PORTER_HOME=" + tmp).Must().RunS()

	return c
}
