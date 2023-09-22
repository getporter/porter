package storage

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/assert"
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

func TestTestParameterProvider_Load(t *testing.T) {
	p := NewTestParameterProvider(t)
	defer p.Close()

	t.Run("unsuccessful load", func(t *testing.T) {
		_, err := p.Load("paramset.json")
		require.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "no such file or directory") || strings.Contains(err.Error(), "The system cannot find the file specified"))
	})

	t.Run("successful load, unsuccessful unmarshal", func(t *testing.T) {
		_, err := p.Load("testdata/paramset_bad.json")
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading testdata/paramset_bad.json as a parameter set")
	})

	t.Run("successful load, successful unmarshal", func(t *testing.T) {
		expected := NewParameterSet("", "mybun",
			secrets.SourceMap{
				Name: "param_env",
				Source: secrets.Source{
					Strategy: "env",
					Hint:     "PARAM_ENV",
				},
			},
			secrets.SourceMap{
				Name: "param_value",
				Source: secrets.Source{
					Strategy: "value",
					Hint:     "param_value",
				},
			},
			secrets.SourceMap{
				Name: "param_command",
				Source: secrets.Source{
					Strategy: "command",
					Hint:     "echo hello world",
				},
			},
			secrets.SourceMap{
				Name: "param_path",
				Source: secrets.Source{
					Strategy: "path",
					Hint:     "/path/to/param",
				},
			},
			secrets.SourceMap{
				Name: "param_secret",
				Source: secrets.Source{
					Strategy: "secret",
					Hint:     "param_secret",
				},
			})
		expected.SchemaVersion = "1.0.1" // It's an older code but it checks out
		expected.Status.Created = time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC)
		expected.Status.Modified = expected.Status.Created

		pset, err := p.Load("testdata/paramset.json")
		require.NoError(t, err)
		require.Equal(t, expected, pset)
	})
}
