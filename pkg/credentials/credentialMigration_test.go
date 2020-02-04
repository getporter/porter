package credentials

import (
	"io/ioutil"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsMigration_ConvertToJson(t *testing.T) {
	c := context.NewTestContext(t)

	oldCS := "/home/.porter/credentials/mybuns.yaml"
	c.AddTestFile("testdata/mybuns.yaml", oldCS)
	lastMod, _ := time.Parse("2006-Jan-02", "2020-Jan-01")
	c.FileSystem.Chtimes(oldCS, lastMod, lastMod)

	m := NewCredentialsMigration(c.Context)
	err := m.ConvertToJson(oldCS)
	require.NoError(t, err, "ConvertToJson failed")

	newCS := "/home/.porter/credentials/mybuns.json"
	exists, _ := c.FileSystem.Exists(newCS)
	assert.True(t, exists, "expected migrated credential set to exist at %s", newCS)
	exists, _ = c.FileSystem.Exists(oldCS)
	assert.True(t, exists, "expected old credential set to still exist at %s", oldCS)

	want, err := ioutil.ReadFile("testdata/mybuns.json")
	require.NoError(t, err, "could not read testdata/mybuns.json")
	got, err := c.FileSystem.ReadFile(newCS)
	require.NoError(t, err, "could not read %s", newCS)
	assert.Equal(t, string(want), string(got))

	require.Len(t, m.migrated, 1, "the wrong number of credential set paths were audited")
	assert.Equal(t, oldCS, m.migrated[0], "invalid credential set path was audited")
}

func TestCredentialsMigration_Migrate_ExistingFileSkipped(t *testing.T) {
	c := context.NewTestContext(t)

	oldCS := "/home/.porter/credentials/mybuns.yaml"
	existingCS := "/home/.porter/credentials/mybuns.json"
	c.AddTestFile("testdata/mybuns.yaml", oldCS)
	c.AddTestFile("testdata/mybuns.json", existingCS)

	fi, _ := c.FileSystem.Stat(existingCS)
	wantMod := fi.ModTime()

	m := NewCredentialsMigration(c.Context)
	err := m.Migrate("/home/.porter/credentials/")
	require.NoError(t, err, "ConvertToJson failed")

	assert.Empty(t, m.migrated, "the credential set should not have been audited as migrated")

	fi, _ = c.FileSystem.Stat(existingCS)
	gotMod := fi.ModTime()
	assert.Equal(t, wantMod, gotMod, "the existing migrated credential set should not have been touched")
}

func TestCredentialsMigration_Migrate(t *testing.T) {
	c := context.NewTestContext(t)

	c.AddTestFile("testdata/mybuns.yaml", "/home/.porter/credentials/bun1.yaml")
	c.AddTestFile("testdata/mybuns.yaml", "/home/.porter/credentials/bun2.yaml")
	c.AddTestFile("testdata/mybuns.json", "/home/.porter/credentials/bun1.json")

	m := NewCredentialsMigration(c.Context)
	err := m.Migrate("/home/.porter/credentials/")
	require.NoError(t, err, "ConvertToJson failed")

	require.Len(t, m.migrated, 1, "the migrated credential sets should have been audited")
	assert.Equal(t, "/home/.porter/credentials/bun2.yaml", m.migrated[0], "invalid credential set path was audited")
}
