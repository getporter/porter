//go:build integration
// +build integration

package integration

import (
	"context"
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
	// TODO(carolynvs): skip until we have migrations defined for 1.0
	t.Skip()
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	ctx := context.Background()

	schema := storage.NewSchema("abc123", "", "")
	p.TestStore.Insert(ctx, migrations.CollectionConfig, storage.InsertOptions{Documents: []interface{}{schema}})

	err := p.MigrateStorage(ctx)
	require.NoError(t, err, "MigrateStorage failed")

	installations, err := p.ListInstallations(ctx, porter.ListOptions{})
	require.NoError(t, err, "could not list installations")
	require.Len(t, installations, 1, "expected one installation")
	assert.Equal(t, "mybun", installations[0].Name, "unexpected list of installation names")
}
