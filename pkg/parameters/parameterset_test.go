package parameters

import (
	"testing"

	"github.com/cnabio/cnab-go/schema"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCNABSpecVersion(t *testing.T) {
	version, err := schema.GetSemver(CNABSpecVersion)
	require.NoError(t, err)
	assert.Equal(t, DefaultSchemaVersion, version)
}

func TestNewCredentialSet(t *testing.T) {
	cs := NewParameterSet("myparams",
		valuesource.Strategy{
			Name: "password",
			Source: valuesource.Source{
				Key:   "env",
				Value: "DB_PASSWORD",
			},
		})

	assert.Equal(t, "myparams", cs.Name, "Name was not set")
	assert.NotEmpty(t, cs.Created, "Created was not set")
	assert.NotEmpty(t, cs.Modified, "Modified was not set")
	assert.Equal(t, cs.Created, cs.Modified, "Created and Modified should have the same timestamp")
	assert.Equal(t, DefaultSchemaVersion, cs.SchemaVersion, "SchemaVersion was not set")
	assert.Len(t, cs.Parameters, 1, "Parameters should be initialized with 1 value")
}
