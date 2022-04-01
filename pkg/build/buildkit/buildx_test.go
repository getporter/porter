package buildkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
