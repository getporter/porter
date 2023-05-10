package storage

import (
	"context"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParameterSet(t *testing.T) {
	ps := NewParameterSet("dev", "myparams",
		secrets.SourceMap{
			Name: "password",
			Source: secrets.Source{
				Strategy: "env",
				Hint:     "DB_PASSWORD",
			},
		})

	assert.Equal(t, DefaultParameterSetSchemaVersion, ps.SchemaVersion, "SchemaVersion was not set")
	assert.Equal(t, "myparams", ps.Name, "Name was not set")
	assert.Equal(t, "dev", ps.Namespace, "Namespace was not set")
	assert.NotEmpty(t, ps.Status.Created, "Created was not set")
	assert.NotEmpty(t, ps.Status.Modified, "Modified was not set")
	assert.Equal(t, ps.Status.Created, ps.Status.Modified, "Created and Modified should have the same timestamp")
	assert.Equal(t, SchemaTypeParameterSet, ps.SchemaType, "incorrect SchemaType")
	assert.Equal(t, DefaultParameterSetSchemaVersion, ps.SchemaVersion, "incorrect SchemaVersion")
	assert.Len(t, ps.Parameters, 1, "Parameters should be initialized with 1 value")
}

func TestParameterSet_String(t *testing.T) {
	t.Run("global namespace", func(t *testing.T) {
		ps := ParameterSet{ParameterSetSpec: ParameterSetSpec{Name: "myparams"}}
		assert.Equal(t, "/myparams", ps.String())
	})

	t.Run("local namespace", func(t *testing.T) {
		ps := ParameterSet{ParameterSetSpec: ParameterSetSpec{Namespace: "dev", Name: "myparams"}}
		assert.Equal(t, "dev/myparams", ps.String())
	})
}

func TestDisplayParameterSet_Validate(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name          string
		schemaType    string
		schemaVersion cnab.SchemaVersion
		wantError     string
	}{
		{
			name:          "schemaType: none",
			schemaType:    "",
			schemaVersion: DefaultParameterSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: ParameterSet",
			schemaType:    SchemaTypeParameterSet,
			schemaVersion: DefaultParameterSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: PARAMETERSET",
			schemaType:    strings.ToUpper(SchemaTypeParameterSet),
			schemaVersion: DefaultParameterSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: parameterset",
			schemaType:    strings.ToUpper(SchemaTypeParameterSet),
			schemaVersion: DefaultParameterSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: CredentialSet",
			schemaType:    SchemaTypeCredentialSet,
			schemaVersion: DefaultParameterSetSchemaVersion,
			wantError:     "invalid schemaType CredentialSet, expected ParameterSet"},
		{
			name:          "validate embedded ps",
			schemaType:    SchemaTypeParameterSet,
			schemaVersion: "", // this is required
			wantError:     "invalid schema version"},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ps := ParameterSet{ParameterSetSpec: ParameterSetSpec{
				SchemaType:    tc.schemaType,
				SchemaVersion: tc.schemaVersion,
			}}
			err := ps.Validate(context.Background(), schema.CheckStrategyExact)
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				tests.RequireErrorContains(t, err, tc.wantError)
			}
		})
	}
}

func TestParameterSet_Validate_DefaultSchemaType(t *testing.T) {
	ps := NewParameterSet("", "myps")
	ps.SchemaType = ""
	require.NoError(t, ps.Validate(context.Background(), schema.CheckStrategyExact))
	assert.Equal(t, SchemaTypeParameterSet, ps.SchemaType)
}
