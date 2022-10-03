package testhelpers

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/carolynvs/magex/pkg/downloads"
	"github.com/carolynvs/magex/shx"
	"github.com/carolynvs/magex/xplat"
	"github.com/stretchr/testify/require"
)

// CreateLegacyPorterHome creates a porter v0.38 PORTER_HOME with legacy data
func CreateLegacyPorterHome(t *testing.T) *config.TestConfig {
	tmp, err := ioutil.TempDir("", "porter")
	require.NoError(t, err)

	c := config.NewTestConfig(t)
	c.DataLoader = config.LoadFromFilesystem()
	c.TestContext.UseFilesystem()
	c.SetHomeDir(tmp)
	c.TestContext.AddCleanupDir(tmp) // Remove the temp home directory when the context is closed

	// Copy old data into PORTER_HOME
	c.TestContext.AddTestDirectoryFromRoot("tests/testdata/porter_home/v0", tmp)

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

	err = c.Load(context.Background(), nil)
	require.NoError(t, err, "Failed to load the test context from the temp PORTER_HOME")
	return c
}
