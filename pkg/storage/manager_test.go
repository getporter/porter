package storage

import (
	"encoding/json"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-go/schema"
	"github.com/cnabio/cnab-go/utils/crud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_LoadSchema(t *testing.T) {
	t.Run("valid schema", func(t *testing.T) {
		schema := Schema{
			Claims:      "cnab-claim-1.0.0-DRAFT",
			Credentials: "cnab-credentials-1.0.0-DRAFT",
		}

		c := config.NewTestConfig(t)
		storage := crud.NewBackingStore(crud.NewMockStore())
		p := NewManager(c.Config, storage)

		schemaB, err := json.Marshal(schema)
		require.NoError(t, err, "Marshal schema failed")
		err = storage.Save("", "", "schema", schemaB)
		require.NoError(t, err, "Save schema failed")

		err = p.loadSchema()
		require.NoError(t, err, "LoadSchema failed")
		assert.NotEmpty(t, p.schema, "Schema should be populated with the file's data")
	})

	t.Run("missing schema", func(t *testing.T) {
		c := config.NewTestConfig(t)
		storage := crud.NewBackingStore(crud.NewMockStore())
		p := NewManager(c.Config, storage)

		err := p.loadSchema()
		require.NoError(t, err, "LoadSchema failed")
		assert.Empty(t, p.schema, "Schema should be empty when none was found")
	})

	t.Run("invalid schema", func(t *testing.T) {
		c := config.NewTestConfig(t)
		storage := crud.NewBackingStore(crud.NewMockStore())
		p := NewManager(c.Config, storage)

		var schemaB = []byte("invalid schema")
		err := storage.Save("", "", "schema", schemaB)
		require.NoError(t, err, "Save schema failed")

		err = p.loadSchema()
		require.Error(t, err, "Expected LoadSchema to fail")
		assert.Contains(t, err.Error(), "could not parse storage schema document")
		assert.Empty(t, p.schema, "Schema should be empty because none was loaded")
	})
}

func TestProvider_ShouldMigrateClaims(t *testing.T) {
	testcases := []struct {
		name         string
		claimVersion string
		wantMigrate  bool
	}{
		{"old schema", "cnab-claim-1.0.0-DRAFT", true},
		{"missing schema", "", true},
		{"current schema", claim.CNABSpecVersion, false},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := config.NewTestConfig(t)
			storage := crud.NewBackingStore(crud.NewMockStore())
			p := NewManager(c.Config, storage)

			p.schema = Schema{
				Claims: schema.Version(tc.claimVersion),
			}

			assert.Equal(t, tc.wantMigrate, p.ShouldMigrateClaims())
		})
	}
}
