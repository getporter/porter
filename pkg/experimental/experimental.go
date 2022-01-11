package experimental

const (
	// BuildDrivers experimental flag
	BuildDrivers   = "build-drivers"
	StructuredLogs = "structured-logs"
)

// FeatureFlags is an enum of possible feature flags
type FeatureFlags int

const (
	// FlagBuildDrivers indicates if configurable build drivers are enabled.
	FlagBuildDrivers FeatureFlags = iota + 1
	FlagStructuredLogs
)

// ParseFlags converts a list of feature flag names into a bit map for faster lookups.
func ParseFlags(flags []string) FeatureFlags {
	var experimental FeatureFlags
	for _, flag := range flags {
		switch flag {
		case BuildDrivers:
			experimental = experimental | FlagBuildDrivers
		case StructuredLogs:
			experimental = experimental | FlagStructuredLogs
		}
	}
	return experimental
}
