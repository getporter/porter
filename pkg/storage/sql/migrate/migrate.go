package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// We can go:embed *.sql as well if we need to.
var embedded embed.FS

func init() { goose.SetBaseFS(embedded) }

func addMigrations(dialect Dialect) {
	m := &migrations{
		Dialect: dialect,
	}
	goose.AddNamedMigrationContext("0001_initial.go", m.Up0001, m.Down0001)
}

type Dialect string

const (
	Postgres Dialect = "postgres"
	// we can easily add more databases here without changing anything in the migration functions
)

func NewDialector(dialect Dialect, tx *sql.Tx, dsn string) (gorm.Dialector, bool) {
	switch dialect {
	case Postgres:
		if tx == nil {
			return postgres.Open(dsn), true
		}
		return postgres.New(postgres.Config{
			Conn: tx,
			DSN:  dsn,
		}), true
	default:
		return nil, false
	}
}

// Migrate runs the migrations for the given dialect and database string.
func Migrate(ctx context.Context, dialect Dialect, dbString string) error {
	if err := goose.SetDialect(string(dialect)); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}
	db, err := newGORM(ctx, dialect, nil, dbString)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	defer sqlDB.Close()

	addMigrations(dialect)
	err = goose.UpContext(ctx, sqlDB, ".")
	if err != nil {
		return fmt.Errorf("failed to migrate up: %w", err)
	}
	return nil
}

func newGORM(ctx context.Context, dialect Dialect, tx *sql.Tx, dsn string) (*gorm.DB, error) {
	dialector, ok := NewDialector(dialect, tx, dsn)
	if !ok {
		return nil, fmt.Errorf("dialect %s not supported", dialect)
	}
	return open(ctx, dialector, &gorm.Config{})
}

func open(ctx context.Context, dialect gorm.Dialector, cfg *gorm.Config) (*gorm.DB, error) {
	db, err := gorm.Open(dialect, cfg)
	if err != nil {
		return nil, err
	}
	return db.WithContext(ctx), nil
}

type migrations struct {
	Dialect Dialect
}

func (m *migrations) GORM(ctx context.Context, tx *sql.Tx) (*gorm.DB, error) {
	return newGORM(ctx, m.Dialect, tx, "")
}
