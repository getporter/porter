package schema

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
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
var ErrInvalidSchemaVersion = errors.New("Invalid schema version")

// ValidateSchemaVersion checks the specified schema version against the supported version,
// returning if the result is a warning only. Warnings are returned when the versions are not an exact match.
// A warning is not returned when CheckStrategyNone is used.
func ValidateSchemaVersion(strategy CheckStrategy, supported string, specified string) (bool, error) {
	if specified == "" {
		specified = "(none)"
	}
	baseMessage := errors.Wrapf(ErrInvalidSchemaVersion, "The schema version is %s but the supported schema version is %s. See https://release-v1.porter.sh/reference/file-formats/#supported-versions for more details.",
		specified, supported)

	switch strategy {
	case CheckStrategyNone:
		// don't return an error ever, but we can still warn down below
	case CheckStrategyExact:
		if specified != supported {
			return false, baseMessage
		}
	case CheckStrategyMinor:
		getMinor := func(version string) string {
			parts := strings.SplitN(version, ".", 3)
			if len(parts) < 2 {
				return version
			}
			return strings.Join(parts[:2], ".")
		}

		specifiedMinor := getMinor(specified)
		supportedMinor := getMinor(supported)

		if specifiedMinor != supportedMinor {
			return false, errors.Wrapf(baseMessage, "The schema version MAJOR.MINOR values do not match")
		}
	case CheckStrategyMajor:
		getMajor := func(version string) string {
			i := strings.Index(version, ".")
			return version[:i]
		}

		specifiedMajor := getMajor(specified)
		supportedMajor := getMajor(supported)

		if specifiedMajor != supportedMajor {
			return false, errors.Wrapf(baseMessage, "The schema version MAJOR values do not match")
		}
	default:
		return false, fmt.Errorf("unknown schema.CheckStrategy %v", strategy)
	}

	if specified == supported {
		return false, nil
	}

	// Even if the check passed, print a warning if it wasn't exactly the same
	return true, errors.Wrap(baseMessage, "WARNING")
}
