//go:build integration
// +build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/storage/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Do a migration. This also checks for any problems with our
// connection handling which can result in panics :-)
func TestClaimMigration_List(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	schema := storage.NewSchema()
	schema.Installations = "v0.38.10"
	p.TestStore.Insert(ctx, migrations.CollectionConfig, storage.InsertOptions{Documents: []interface{}{schema}})

	opts := porter.MigrateStorageOptions{
		Source:      "TODO",
		Destination: "TODO",
	}
	err := p.MigrateStorage(ctx, opts)
	require.NoError(t, err, "MigrateStorage failed")

	installations, err := p.ListInstallations(ctx, porter.ListOptions{})
	require.NoError(t, err, "could not list installations")
	require.Len(t, installations, 1, "expected one installation")
	assert.Equal(t, "mybun", installations[0].Name, "unexpected list of installation names")
}
