package migrate

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/robinbraemer/devroach"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMigrate(t *testing.T) {
	d := devroach.NewPoolT(t, nil)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	db, err := newGORM(ctx, Postgres, nil, d.Config().ConnString())
	require.NoError(t, err)

	err = MigrateDB(ctx, db)
	require.NoError(t, err)

	checkTables(t, db, []string{
		"goose_db_version",
		"installations",
		"outputs",
		"runs",
		"results",
		"credential_sets",
		"parameter_sets",
	})

	err = migrateDown(ctx, db)
	require.NoError(t, err)

	checkTables(t, db, []string{"goose_db_version"})
}

func checkTables(t *testing.T, db *gorm.DB, tables []string) {
	t.Helper()

	actualTables, err := db.Migrator().GetTables()
	require.NoError(t, err)

	sort.Strings(actualTables)
	sort.Strings(tables)
	require.EqualValues(t, tables, actualTables)
}
