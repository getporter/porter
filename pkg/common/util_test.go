package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRandomString(t *testing.T) {
	for _, test := range []struct {
		seed     string
		length   int
		expected string
	}{
		{"", 10, "IC9qGbVw9G"},
		{"getporter/porter-hello", 8, "FVRaTSxR"},
		{"getporter/porter-hello:1.0.0", 10, "vshypaGk9Z"},
	} {
		result := RandomString(test.seed, test.length)
		require.Equalf(t, test.expected, result, "failed with seed %s", test.seed)
	}
}
