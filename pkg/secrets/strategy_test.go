package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet_Merge(t *testing.T) {
	set := Set{
		"first":  "first",
		"second": "second",
		"third":  "third",
	}

	is := assert.New(t)

	err := set.Merge(Set{})
	is.NoError(err)
	is.Len(set, 3)
	is.NotContains(set, "fourth")

	err = set.Merge(Set{"fourth": "fourth"})
	is.NoError(err)
	is.Len(set, 4)
	is.Contains(set, "fourth")

	err = set.Merge(Set{"second": "bis"})
	is.EqualError(err, `ambiguous value resolution: "second" is already present in base sets, cannot merge`)
}

func TestSource_UnmarshalRaw(t *testing.T) {
	tests := []struct {
		name string
		raw  map[string]interface{}
		want Source
		err  string
	}{
		{
			name: "empty map",
			raw:  map[string]interface{}{},
			want: Source{},
		},
		{
			name: "string",
			raw: map[string]interface{}{
				"env": "SOME_VALUE",
			},
			want: Source{
				Strategy: "env",
				Hint:     "SOME_VALUE",
			},
		},
		{
			name: "array value",
			raw: map[string]interface{}{
				"value": []interface{}{1, 2, "3"},
			},
			want: Source{
				Strategy: "value",
				Hint:     "[1,2,\"3\"]",
			},
		},
		{
			name: "map value",
			raw: map[string]interface{}{
				"value": map[string]interface{}{
					"abc": "def",
				},
			},
			want: Source{
				Strategy: "value",
				Hint:     `{"abc":"def"}`,
			},
		},
		{
			name: "integer value",
			raw: map[string]interface{}{
				"value": 10,
			},
			want: Source{
				Strategy: "value",
				Hint:     "10",
			},
		},
		{
			name: "float value",
			raw: map[string]interface{}{
				"value": 3.1415,
			},
			want: Source{
				Strategy: "value",
				Hint:     "3.1415",
			},
		},
		{
			name: "boolean value",
			raw: map[string]interface{}{
				"value": true,
			},
			want: Source{
				Strategy: "value",
				Hint:     "true",
			},
		},
		{
			name: "null value",
			raw: map[string]interface{}{
				"value": nil,
			},
			want: Source{
				Strategy: "value",
				Hint:     "",
			},
		},
		{
			name: "multiple keys",
			raw: map[string]interface{}{
				"env":   "abc",
				"value": "def",
			},
			err: "multiple key/value pairs specified for source but only one may be defined",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := Source{}
			err := s.UnmarshalRaw(tc.raw)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			assert.Equal(t, tc.want, s)
		})
	}
}
