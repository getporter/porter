package tracing

import "testing"

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
	} {
		fn, ok := extractFuncName(test.input)
		if fn != test.expected || ok != test.ok {
			t.Errorf("failed %q, got %q %v, expected %q %v", test.input, fn, ok, test.expected, test.ok)
		}
	}
}
