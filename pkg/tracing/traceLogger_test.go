package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFuncName(t *testing.T) {
	for _, test := range []struct {
		input    string
		expected string
		ok       bool
	}{
		{"", "", false},
		{"porter/", "", false},
		{"porter/v.", "", false},
		{"github.com/getporter/porter/tracing.StartSpan", "StartSpan", true},
		{"get.porter.sh/porter/pkg/porter.(*Porter).ListInstallations", "ListInstallations", true},
	} {
		fn, ok := extractFuncName(test.input)
		assert.Equal(t, test.expected, fn, "failed with input %q", test.input)
		assert.Equal(t, test.ok, ok, "failed with input %q", test.input)
	}
}
