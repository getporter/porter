package parameters

import (
	"testing"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/assert"
)

func TestNewParameterSet(t *testing.T) {
	cs := NewParameterSet("dev", "myparams",
		secrets.Strategy{
			Name: "password",
			Source: secrets.Source{
				Key:   "env",
				Value: "DB_PASSWORD",
			},
		})

	assert.Equal(t, "myparams", cs.Name, "Name was not set")
	assert.Equal(t, "dev", cs.Namespace, "Namespace was not set")
	assert.NotEmpty(t, cs.Created, "Created was not set")
	assert.NotEmpty(t, cs.Modified, "Modified was not set")
	assert.Equal(t, cs.Created, cs.Modified, "Created and Modified should have the same timestamp")
	assert.Equal(t, SchemaVersion, cs.SchemaVersion, "SchemaVersion was not set")
	assert.Len(t, cs.Parameters, 1, "Parameters should be initialized with 1 value")
}
