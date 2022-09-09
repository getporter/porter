package storage

import (
	"testing"

	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParameterSet(t *testing.T) {
	ps := NewParameterSet("dev", "myparams",
		secrets.Strategy{
			Name: "password",
			Source: secrets.Source{
				Key:   "env",
				Value: "DB_PASSWORD",
			},
		})

	assert.Equal(t, ParameterSetSchemaVersion, ps.SchemaVersion, "SchemaVersion was not set")
	assert.Equal(t, "myparams", ps.Name, "Name was not set")
	assert.Equal(t, "dev", ps.Namespace, "Namespace was not set")
	assert.NotEmpty(t, ps.Status.Created, "Created was not set")
	assert.NotEmpty(t, ps.Status.Modified, "Modified was not set")
	assert.Equal(t, ps.Status.Created, ps.Status.Modified, "Created and Modified should have the same timestamp")
	assert.Equal(t, ParameterSetSchemaVersion, ps.SchemaVersion, "SchemaVersion was not set")
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
		schemaVersion schema.Version
		wantError     string
	}{
		{
			name:          "schemaType: none",
			schemaType:    "",
			schemaVersion: ParameterSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: ParameterSet",
			schemaType:    "ParameterSet",
			schemaVersion: ParameterSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: PARAMETERSET",
			schemaType:    "PARAMETERSET",
			schemaVersion: ParameterSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: parameterset",
			schemaType:    "parameterset",
			schemaVersion: ParameterSetSchemaVersion,
			wantError:     ""},
		{
			name:          "schemaType: CredentialSet",
			schemaType:    "CredentialSet",
			schemaVersion: ParameterSetSchemaVersion,
			wantError:     "invalid schemaType CredentialSet, expected ParameterSet"},
		{
			name:          "validate embedded ps",
			schemaType:    "ParameterSet",
			schemaVersion: "", // this is required
			wantError:     "invalid schemaVersion provided: (none)"},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ps := ParameterSet{ParameterSetSpec: ParameterSetSpec{
				SchemaType:    tc.schemaType,
				SchemaVersion: tc.schemaVersion,
			}}
			err := ps.Validate()
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				tests.RequireErrorContains(t, err, tc.wantError)
			}
		})
	}
}
