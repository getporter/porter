package schema

import (
	"testing"

	"get.porter.sh/porter/tests"
	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSchemaVersion(t *testing.T) {
	testcases := []struct {
		Name        string
		Strategy    CheckStrategy
		Supported   string
		Default     string
		Specified   string
		WantWarning bool
		WantErr     string
	}{
		{Name: "range - in range", Strategy: CheckStrategyExact, Supported: "1.0.0-alpha.1 - 1.0.0-z || 1.0.0 - 1.0.1", Specified: "1.0.0-alpha.2", Default: "1.0.1"},
		{Name: "range - out of range", Strategy: CheckStrategyExact, Supported: "1.0.0-alpha.1 - 1.0.1", Specified: "1.0.2", Default: "1.0.1",
			WantErr: "the schema version is 1.0.2 but the supported schema version is >=1.0.0-alpha.1 <=1.0.1"},
		{Name: "exact - exact match", Strategy: CheckStrategyExact, Supported: "=1.0.0", Specified: "1.0.0", Default: "1.0.0"},
		{Name: "exact - patch mismatch", Strategy: CheckStrategyExact, Supported: "=1.0.0", Specified: "1.0.1", Default: "1.0.0",
			WantErr: "the schema version is 1.0.1 but the supported schema version is =1.0.0"},
		{Name: "minor - exact match", Strategy: CheckStrategyMinor, Supported: "=1.1.2", Specified: "1.1.2", Default: "1.1.2"},
		{Name: "minor - minor match", Strategy: CheckStrategyMinor, Supported: "=1.1.2", Specified: "1.1.1", Default: "1.1.2",
			WantWarning: true, WantErr: "WARNING: the schema version is 1.1.1 but the supported schema version is =1.1.2"},
		{Name: "minor - minor mismatch", Strategy: CheckStrategyMinor, Supported: "=1.2.0", Specified: "1.1.0", Default: "1.2.0",
			WantErr: "the schema version MAJOR.MINOR values do not match"},
		{Name: "major - exact match", Strategy: CheckStrategyMajor, Supported: "=1.1.2", Specified: "1.1.2", Default: "1.2.0"},
		{Name: "major - major match", Strategy: CheckStrategyMajor, Supported: "=1.1.2", Specified: "1.1.1", Default: "1.1.2",
			WantWarning: true, WantErr: "WARNING: the schema version is 1.1.1 but the supported schema version is =1.1.2"},
		{Name: "major - major mismatch", Strategy: CheckStrategyMajor, Supported: "=1.1.2", Specified: "2.0.0", Default: "1.1.2",
			WantErr: "the schema version MAJOR values do not match"},
		{Name: "none - mismatch", Strategy: CheckStrategyNone, Supported: "=1.0.0", Specified: "2.0.0", Default: "1.0.0",
			WantWarning: true, WantErr: "WARNING: the schema version is 2.0.0 but the supported schema version is =1.0.0"},
		{Name: "none - invalid version", Strategy: CheckStrategyNone, Supported: "=1.0.0", Specified: "", Default: "1.0.0",
			WantWarning: true, WantErr: "(none) is not a valid semantic version"},
	}
	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			supportedC, err := semver.NewConstraint(tc.Supported)
			require.NoError(t, err, "Failed to parse supported constraint")
			defaultV, err := semver.NewVersion(tc.Default)
			require.NoError(t, err, "Failed to parse default version")

			gotWarning, err := ValidateSchemaVersion(tc.Strategy, supportedC, tc.Specified, defaultV)
			if tc.WantErr != "" {
				tests.RequireErrorContains(t, err, tc.WantErr, "incorrect message returned")
			} else {
				require.NoError(t, err, "ValidateSchemaVersion failed")
			}

			assert.Equal(t, tc.WantWarning, gotWarning, "unexpected warning result")
		})
	}
}
