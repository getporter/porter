package parameters

import (
	"reflect"
	"testing"
	"time"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/valuesource"
	"github.com/stretchr/testify/require"
)

func TestParseVariableAssignments(t *testing.T) {
	testcases := []struct {
		Name, Raw, Variable, Value string
	}{
		{"simple", "a=b", "a", "b"},
		{"multiple equal signs", "c=abc1232===", "c", "abc1232==="},
		{"empty value", "d=", "d", ""},
		{"extra whitespace", " a = b ", "a", "b"},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {

			params := []string{tc.Raw}

			got, err := ParseVariableAssignments(params)
			if err != nil {
				t.Fatal(err)
			}

			want := make(map[string]string)
			want[tc.Variable] = tc.Value
			if !reflect.DeepEqual(want, got) {
				t.Fatalf("%s\nexpected:\n\t%v\ngot:\n\t%v\n", tc.Raw, want, got)
			}
		})
	}
}

func TestParseVariableAssignments_MissingVariableName(t *testing.T) {
	params := []string{"=b"}

	_, err := ParseVariableAssignments(params)
	if err == nil {
		t.Fatal("should have failed due to a missing variable name")
	}
}

func TestLoad(t *testing.T) {
	t.Run("unsuccessful load", func(t *testing.T) {
		_, err := Load("paramset.json")
		require.EqualError(t, err, "open paramset.json: no such file or directory")
	})

	t.Run("successful load, unsuccessful unmarshal", func(t *testing.T) {
		expected := &ParameterSet{}

		pset, err := Load("testdata/paramset_bad.json")
		require.EqualError(t, err,
			"yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `myparam...` into parameters.ParameterSet")
		require.Equal(t, expected, pset)
	})

	t.Run("successful load, successful unmarshal", func(t *testing.T) {
		expected := &ParameterSet{
			Name:     "mybun",
			Created:  time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC),
			Modified: time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC),
			Parameters: []valuesource.Strategy{
				{
					Name: "param_env",
					Source: valuesource.Source{
						Key:   "env",
						Value: "PARAM_ENV",
					},
				},
				{
					Name: "param_value",
					Source: valuesource.Source{
						Key:   "value",
						Value: "param_value",
					},
				},
				{
					Name: "param_command",
					Source: valuesource.Source{
						Key:   "command",
						Value: "echo hello world",
					},
				},
				{
					Name: "param_path",
					Source: valuesource.Source{
						Key:   "path",
						Value: "/path/to/param",
					},
				},
				{
					Name: "param_secret",
					Source: valuesource.Source{
						Key:   "secret",
						Value: "param_secret",
					},
				},
			},
		}

		pset, err := Load("testdata/paramset.json")
		require.NoError(t, err)
		require.Equal(t, expected, pset)
	})
}

func TestIsInternal(t *testing.T) {
	bun := bundle.Bundle{
		Definitions: definition.Definitions{
			"foo": &definition.Schema{
				Type: "string",
			},
			"porter-debug": &definition.Schema{
				Type:    "string",
				Comment: PorterInternal,
			},
		},
		Parameters: map[string]bundle.Parameter{
			"foo": {
				Definition: "foo",
			},
			"baz": {
				Definition: "baz",
			},
			"porter-debug": {
				Definition: "porter-debug",
			},
		},
	}

	t.Run("empty bundle", func(t *testing.T) {
		require.False(t, IsInternal("foo", bundle.Bundle{}))
	})

	t.Run("param does not exist", func(t *testing.T) {
		require.False(t, IsInternal("bar", bun))
	})

	t.Run("definition does not exist", func(t *testing.T) {
		require.False(t, IsInternal("baz", bun))
	})

	t.Run("is not internal", func(t *testing.T) {
		require.False(t, IsInternal("foo", bun))
	})

	t.Run("is internal", func(t *testing.T) {
		require.True(t, IsInternal("porter-debug", bun))
	})
}
