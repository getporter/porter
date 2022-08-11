package storage

import (
	"encoding/json"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/test"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_NewResultFrom(t *testing.T) {
	run := NewRun("dev", "mybuns")
	cnabResult := cnab.Result{
		ID:             "resultID",
		ClaimID:        "claimID",
		Created:        time.Now(),
		Message:        "message",
		Status:         "status",
		OutputMetadata: cnab.OutputMetadata{"myoutput": map[string]string{}},
		Custom:         map[string]interface{}{"custom": true},
	}

	result := run.NewResultFrom(cnabResult)
	assert.Equal(t, cnabResult.ID, result.ID)
	assert.Equal(t, run.Namespace, result.Namespace)
	assert.Equal(t, run.Installation, result.Installation)
	assert.Equal(t, run.ID, result.RunID)
	assert.Equal(t, cnabResult.Created, result.Created)
	assert.Equal(t, cnabResult.Status, result.Status)
	assert.Equal(t, cnabResult.Message, result.Message)
	assert.Equal(t, cnabResult.OutputMetadata, result.OutputMetadata)
	assert.Equal(t, cnabResult.Custom, result.Custom)
}

func TestRun_ShouldRecord(t *testing.T) {
	t.Run("stateless, not modifies", func(t *testing.T) {
		b := bundle.Bundle{
			Actions: map[string]bundle.Action{
				"dry-run": {
					Modifies:  false,
					Stateless: true,
				},
			},
		}

		r := Run{Bundle: b, Action: "dry-run"}
		assert.False(t, r.ShouldRecord())
	})

	t.Run("stateful, not modifies", func(t *testing.T) {
		b := bundle.Bundle{
			Actions: map[string]bundle.Action{
				"audit": {
					Modifies:  false,
					Stateless: false,
				},
			},
		}

		r := Run{Bundle: b, Action: "audit"}
		assert.True(t, r.ShouldRecord())
	})

	t.Run("modifies", func(t *testing.T) {
		b := bundle.Bundle{
			Actions: map[string]bundle.Action{
				"editstuff": {
					Modifies:  true,
					Stateless: false,
				},
			},
		}

		r := Run{Bundle: b, Action: "editstuff"}
		assert.True(t, r.ShouldRecord())
	})

	t.Run("missing definition", func(t *testing.T) {
		b := bundle.Bundle{}

		r := Run{Bundle: b, Action: "missing"}
		assert.True(t, r.ShouldRecord())
	})

	t.Run("has user defined output", func(t *testing.T) {
		b := bundle.Bundle{
			Actions: map[string]bundle.Action{
				"editstuff": {
					Modifies:  false,
					Stateless: true,
				},
			},
			Outputs: map[string]bundle.Output{
				"testdata": {
					ApplyTo: []string{"editstuff"},
				},
			},
		}

		r := Run{Bundle: b, Action: "editstuff"}
		assert.True(t, r.ShouldRecord())
	})

	t.Run("has only internal bundle level output", func(t *testing.T) {
		b := bundle.Bundle{
			Definitions: definition.Definitions{
				"porter-state": &definition.Schema{
					Type:            "string",
					ContentEncoding: "base64",
					Comment:         cnab.PorterInternal,
				},
			},
			Actions: map[string]bundle.Action{
				"editstuff": {
					Modifies:  false,
					Stateless: true,
				},
			},
			Outputs: map[string]bundle.Output{
				"porter-state": {Definition: "porter-state"},
			},
		}

		r := Run{Bundle: b, Action: "editstuff"}
		assert.False(t, r.ShouldRecord())
	})

}

func TestRun_TypedParameterValues(t *testing.T) {
	sensitive := true
	bun := bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type:      "integer",
				WriteOnly: &sensitive,
			},
			"baz": &definition.Schema{
				Type: "string",
			},
			"porter-state": &definition.Schema{
				Type:            "string",
				ContentEncoding: "base64",
				Comment:         cnab.PorterInternal,
			},
		},
		Parameters: map[string]bundle.Parameter{
			"foo": {
				Definition: "foo",
			},
			"baz": {
				Definition: "baz",
			},
			"name": {
				Definition: "name",
			},
			"porter-state": {
				Definition: "porter-state",
			},
		},
		RequiredExtensions: []string{
			cnab.FileParameterExtensionKey,
		},
	}

	run := NewRun("dev", "mybuns")
	run.Bundle = bun
	run.Parameters = NewParameterSet(run.Namespace, run.Bundle.Name)
	params := []secrets.Strategy{
		ValueStrategy("baz", "baz-test"),
		ValueStrategy("name", "porter-test"),
		ValueStrategy("porter-state", ""),
		{Name: "foo", Source: secrets.Source{Key: secrets.SourceSecret, Value: "runID"}, Value: "5"},
	}

	expected := map[string]interface{}{
		"baz":          "baz-test",
		"name":         "porter-test",
		"porter-state": nil,
		"foo":          5,
	}

	run.Parameters.Parameters = params
	typed := run.TypedParameterValues()
	require.Equal(t, len(params), len(typed))
	require.Equal(t, len(expected), len(typed))

	for name, value := range typed {
		v, ok := expected[name]
		require.True(t, ok)
		require.Equal(t, v, value)
	}
}

func TestRun_MarshalJSON(t *testing.T) {
	// Verify that when a run is marshaled that the bundle field is saved as an escaped json string
	r1 := Run{ID: "foo", Bundle: exampleBundle}

	data, err := json.Marshal(r1)
	require.NoError(t, err, "Marshal failed")

	test.CompareGoldenFile(t, "testdata/marshaled_run.json", string(data))

	var r2 Run
	err = json.Unmarshal(data, &r2)
	require.NoError(t, err, "Unmarshal failed")

	assert.Equal(t, r1, r2, "The run did not survive the round trip")
}
