package schema

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// MustParseConstraint converts the string value to a semver range. This will panic if it is not a valid constraint and is intended to initialize package variables with well-known schema version values.
// Example:
// var SupportedInstallationSchemaVersion = schema.MustParseConstraint("1.0.x")
func MustParseConstraint(value string) *semver.Constraints {
	c, err := semver.NewConstraint(value)
	if err != nil {
		panic(fmt.Errorf("invalid semver constraint %s: %w", value, err))
	}
	return c
}
