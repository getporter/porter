package migrate

import (
	"context"
	"testing"
	"time"

	"github.com/robinbraemer/devroach"
	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	d := devroach.NewPoolT(t, nil)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	err := Migrate(ctx, Postgres, d.Config().ConnString())
	require.NoError(t, err)

	db, err := newGORM(ctx, Postgres, nil, d.Config().ConnString())
	require.NoError(t, err)

	tables, err := db.Migrator().GetTables()
	require.NoError(t, err)
	require.Greater(t, len(tables), 2)
}
