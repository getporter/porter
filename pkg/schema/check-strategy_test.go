package schema

import (
	"testing"

	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
)

func TestValidateSchemaVersion(t *testing.T) {
	testcases := []struct {
		name        string
		strategy    CheckStrategy
		supported   string
		specified   string
		wantWarning bool
		wantErr     string
	}{
		{name: "exact - exact match", strategy: CheckStrategyExact, supported: "1.0.0", specified: "1.0.0"},
		{name: "exact - patch mismatch", strategy: CheckStrategyExact, supported: "1.0.0", specified: "1.0.1",
			wantErr: "the schema version is 1.0.1 but the supported schema version is 1.0.0"},
		{name: "minor - exact match", strategy: CheckStrategyMinor, supported: "1.1.2", specified: "1.1.2"},
		{name: "minor - minor match", strategy: CheckStrategyMinor, supported: "1.1.2", specified: "1.1.1",
			wantWarning: true, wantErr: "WARNING: the schema version is 1.1.1 but the supported schema version is 1.1.2"},
		{name: "minor - minor mismatch", strategy: CheckStrategyMinor, supported: "1.2.0", specified: "1.1.0",
			wantErr: "the schema version MAJOR.MINOR values do not match"},
		{name: "major - exact match", strategy: CheckStrategyMajor, supported: "1.1.2", specified: "1.1.2"},
		{name: "major - major match", strategy: CheckStrategyMajor, supported: "1.1.2", specified: "1.1.1",
			wantWarning: true, wantErr: "WARNING: the schema version is 1.1.1 but the supported schema version is 1.1.2"},
		{name: "major - major mismatch", strategy: CheckStrategyMajor, supported: "1.1.2", specified: "2.0.0",
			wantErr: "the schema version MAJOR values do not match"},
		{name: "none - mismatch", strategy: CheckStrategyNone, supported: "1.0.0", specified: "2.0.0",
			wantWarning: true, wantErr: "WARNING: the schema version is 2.0.0 but the supported schema version is 1.0.0"},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotWarning, err := ValidateSchemaVersion(tc.strategy, tc.supported, tc.specified)
			if tc.wantErr != "" {
				tests.RequireErrorContains(t, err, tc.wantErr, "incorrect message returned")
			}

			assert.Equal(t, tc.wantWarning, gotWarning, "unexpected warning result")
		})
	}
}
