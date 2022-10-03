package buildkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseBuildArgs(t *testing.T) {
	testcases := []struct {
		name      string
		inputArgs []string
		wantArgs  map[string]string
	}{
		{name: "valid args", inputArgs: []string{"A=1", "B=2=2", "C="},
			wantArgs: map[string]string{"A": "1", "B": "2=2", "C": ""}},
		{name: "missing equal sign", inputArgs: []string{"A"},
			wantArgs: map[string]string{}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var gotArgs = map[string]string{}
			parseBuildArgs(tc.inputArgs, gotArgs)
			assert.Equal(t, tc.wantArgs, gotArgs)
		})
	}
}

func Test_flattenMap(t *testing.T) {
	tt := []struct {
		desc string
		inp  map[string]interface{}
		out  map[string]string
		err  bool
	}{
		{
			desc: "one pair",
			inp: map[string]interface{}{
				"key": "value",
			},
			out: map[string]string{
				"key": "value",
			},
			err: false,
		},
		{
			desc: "nested input",
			inp: map[string]interface{}{
				"key": map[string]string{
					"nestedKey": "value",
				},
			},
			out: map[string]string{
				"key.nestedKey": "value",
			},
			err: false,
		},
		{
			desc: "nested input",
			inp: map[string]interface{}{
				"key1": map[string]interface{}{
					"key2": map[string]string{
						"key3": "value",
					},
				},
			},
			out: map[string]string{
				"key1.key2.key3": "value",
			},
			err: false,
		},
		{
			desc: "multiple nested input",
			inp: map[string]interface{}{
				"key11": map[string]interface{}{
					"key12": map[string]string{
						"key13": "value1",
					},
				},
				"key21": map[string]string{
					"key22": "value2",
				},
			},
			out: map[string]string{
				"key11.key12.key13": "value1",
				"key21.key22":       "value2",
			},
			err: false,
		},
		{
			// CNAB represents null parameters as empty strings, so we will do the same, e.g. ARG CUSTOM_FOO=
			desc: "nil is converted empty string",
			inp: map[string]interface{}{
				"a": nil,
			},
			out: map[string]string{
				"a": "",
			},
			err: false,
		},
		{
			desc: "int is converted to string representation",
			inp: map[string]interface{}{
				"a": 1,
			},
			out: map[string]string{
				"a": "1",
			},
			err: false,
		},
		{
			desc: "bool is converted to string representation",
			inp: map[string]interface{}{
				"a": true,
			},
			out: map[string]string{
				"a": "true",
			},
			err: false,
		},
		{
			desc: "array is converted to string representation",
			inp: map[string]interface{}{
				"a": []string{"beep", "boop"},
			},
			out: map[string]string{
				"a": `["beep","boop"]`,
			},
			err: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			out, err := flattenMap(tc.inp)
			if tc.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.out, out)
		})
	}
}
