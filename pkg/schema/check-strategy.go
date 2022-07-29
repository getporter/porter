package schema

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// CheckStrategy is an enum of values for handling schemaVersion
// comparisons of two resources. Allowed values are: CheckStrategyExact,
// CheckStrategyMinor, CheckStrategyMajor, CheckStrategyNone.
type CheckStrategy string

const (
	// CheckStrategyExact requires that resource schemaVersion values exactly match the supported schema version.
	CheckStrategyExact CheckStrategy = "exact"

	// CheckStrategyMinor requires that resource schemaVersion values match the MAJOR.MINOR portion of the supported schema version.
	CheckStrategyMinor CheckStrategy = "minor"

	// CheckStrategyMajor requires that resource schemaVersion values exactly match the MAJOR portion of the supported schema version.
	CheckStrategyMajor CheckStrategy = "major"

	// CheckStrategyNone ignores the resource schemaVersion. Errors will most likely ensue but have fun!
	CheckStrategyNone CheckStrategy = "none"
)

// ErrInvalidSchemaVersion is used when the schemaVersion of two resources do not match exactly.
var ErrInvalidSchemaVersion = errors.New("invalid schema version")

// ValidateSchemaVersion checks the specified schema version against the supported version,
// returning if the result is a warning only. Warnings are returned when the versions are not an exact match.
// A warning is not returned when CheckStrategyNone is used.
func ValidateSchemaVersion(strategy CheckStrategy, supported *semver.Constraints, specified string, defaultVersion *semver.Version) (bool, error) {
	if specified == "" {
		specified = "(none)"
	}
	baseMessage := fmt.Errorf("the schema version is %s but the supported schema version is %s. See https://getporter.org/reference/file-formats/#supported-versions for more details: %w",
		specified, supported, ErrInvalidSchemaVersion)

	specifiedV, err := semver.NewVersion(specified)
	if err != nil {
		isWarning := strategy == CheckStrategyNone
		return isWarning, fmt.Errorf("%s is not a valid semantic version: %w", specified, ErrInvalidSchemaVersion)
	}

	isSchemaVersionSatisfied := supported.Check(specifiedV)
	switch strategy {
	case CheckStrategyNone:
		// this strategy always passes
	case CheckStrategyExact:
		// Check if the schema version satisfies the supported version range
		if isSchemaVersionSatisfied {
			return false, nil
		} else {
			return false, baseMessage
		}
	case CheckStrategyMinor:
		// Check if the schema version matches the MAJOR.MINOR version number of the currently supported (default) schema version
		supportedMinor, _ := semver.NewConstraint(fmt.Sprintf("~%d.%d.0-0", defaultVersion.Major(), defaultVersion.Minor()))
		isMinorMatch := supportedMinor.Check(specifiedV)
		if !isMinorMatch {
			return false, fmt.Errorf("the schema version MAJOR.MINOR values do not match: %w", baseMessage)
		}
	case CheckStrategyMajor:
		// Check if the schema version matches the MAJOR version number of the currently supported (default) schema version
		supportedMajor, _ := semver.NewConstraint(fmt.Sprintf("^%d.0.0-0", defaultVersion.Major()))
		isMajorMatch := supportedMajor.Check(specifiedV)
		if !isMajorMatch {
			return false, fmt.Errorf("the schema version MAJOR values do not match: %w", baseMessage)
		}
	default:
		return false, fmt.Errorf("unknown schema.CheckStrategy %v", strategy)
	}

	if isSchemaVersionSatisfied {
		return false, nil
	} else {
		// Even if the check passed, print a warning if it wasn't strictly supported
		return true, fmt.Errorf("WARNING: %w", baseMessage)
	}
}
