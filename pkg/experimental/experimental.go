package experimental

const (
	StructuredLogs = "structured-logs"
)

// FeatureFlags is an enum of possible feature flags
type FeatureFlags int

const (
	// FlagStructuredLogs indicates if structured logs are enabled
	FlagStructuredLogs FeatureFlags = iota + 1
)

// ParseFlags converts a list of feature flag names into a bit map for faster lookups.
func ParseFlags(flags []string) FeatureFlags {
	var experimental FeatureFlags
	for _, flag := range flags {
		switch flag {
		case StructuredLogs:
			experimental = experimental | FlagStructuredLogs
		}
	}
	return experimental
}
