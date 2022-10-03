package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseLogLevel(t *testing.T) {
	testcases := []struct {
		value string
		want  LogLevel
	}{
		{"debug", LogLevelDebug},
		{"info", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"warning", LogLevelWarn}, // Allow for our humanity
		{"error", LogLevelError},
	}
	for _, tc := range testcases {
		t.Run(tc.value, func(t *testing.T) {
			lvl := ParseLogLevel(tc.value)
			assert.Equalf(t, tc.want, lvl, "ParseLogLevel(%v)", tc.value)
		})
	}
}
