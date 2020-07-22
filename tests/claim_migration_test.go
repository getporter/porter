// +build integration

package tests

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Do a transparent migration. This also checks for any problems with our
// connection handling which can result in panics :-)
func TestClaimMigration_List(t *testing.T) {
	p := porter.NewTestPorter(t)
	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()

	// Make a test home
	home, err := p.GetHomeDir()
	require.NoError(t, err, "GetHomeDir failed")
	claimsDir := filepath.Join(home, "claims")

	// Remove any rando stuff copied from the dev bin, you won't find this in CI but a local dev run may have it
	err = p.FileSystem.RemoveAll(claimsDir)
	require.NoError(t, err, "error removing existing claims directory before test run")
	err = p.FileSystem.Remove(filepath.Join(home, "schema.json"))
	require.NoError(t, err, "error removing existing schema.json")

	// Create unmigrated claim data
	p.FileSystem.Mkdir(claimsDir, 0755)
	p.AddTestFile(filepath.Join("../pkg/claims/testdata", "upgraded.json"), filepath.Join(home, "claims", "mybun.json"))

	installations, err := p.ListInstallations()
	require.NoError(t, err, "could not list installations")
	require.Len(t, installations, 1, "expected one installation")
	assert.Equal(t, "mybun", installations[0].Name, "unexpected list of installation names")
}
