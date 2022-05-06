package storage

import (
	"reflect"
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
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("successful load, unsuccessful unmarshal", func(t *testing.T) {
		_, err := p.Load("testdata/paramset_bad.json")
		require.Error(t, err)
		require.Contains(t, err.Error(), "error reading testdata/paramset_bad.json as a parameter set")
	})

	t.Run("successful load, successful unmarshal", func(t *testing.T) {
		expected := NewParameterSet("", "mybun",
			secrets.Strategy{
				Name: "param_env",
				Source: secrets.Source{
					Key:   "env",
					Value: "PARAM_ENV",
				},
			},
			secrets.Strategy{
				Name: "param_value",
				Source: secrets.Source{
					Key:   "value",
					Value: "param_value",
				},
			},
			secrets.Strategy{
				Name: "param_command",
				Source: secrets.Source{
					Key:   "command",
					Value: "echo hello world",
				},
			},
			secrets.Strategy{
				Name: "param_path",
				Source: secrets.Source{
					Key:   "path",
					Value: "/path/to/param",
				},
			},
			secrets.Strategy{
				Name: "param_secret",
				Source: secrets.Source{
					Key:   "secret",
					Value: "param_secret",
				},
			})
		expected.ID = "abc123"
		expected.Status.Created = time.Date(1983, time.April, 18, 1, 2, 3, 4, time.UTC)
		expected.Status.Modified = expected.Status.Created

		pset, err := p.Load("testdata/paramset.json")
		require.NoError(t, err)
		require.Equal(t, expected, pset)
	})
}
