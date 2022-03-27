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

	assert.Equal(t, SchemaVersion, cs.SchemaVersion, "SchemaVersion was not set")
	assert.Equal(t, "myparams", cs.Name, "Name was not set")
	assert.Equal(t, "dev", cs.Namespace, "Namespace was not set")
	assert.NotEmpty(t, cs.Status.Created, "Created was not set")
	assert.NotEmpty(t, cs.Status.Modified, "Modified was not set")
	assert.Equal(t, cs.Status.Created, cs.Status.Modified, "Created and Modified should have the same timestamp")
	assert.Equal(t, SchemaVersion, cs.SchemaVersion, "SchemaVersion was not set")
	assert.Len(t, cs.Parameters, 1, "Parameters should be initialized with 1 value")
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
