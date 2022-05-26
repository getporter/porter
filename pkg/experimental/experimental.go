package experimental

const (
	// NoopFeature is a placeholder feature flag that allows us to test our feature flag functions even when there are no active feature flags
	NoopFeature = "no-op"
)

// FeatureFlags is an enum of possible feature flags
type FeatureFlags int

const (
	// FlagNoopFeature is a placeholder feature flag that allows us to test our feature flag functions even when there are no active feature flags
	FlagNoopFeature FeatureFlags = iota + 1
)

// ParseFlags converts a list of feature flag names into a bit map for faster lookups.
func ParseFlags(flags []string) FeatureFlags {
	var experimental FeatureFlags
	for _, flag := range flags {
		switch flag {
		case NoopFeature:
			experimental = experimental | FlagNoopFeature
		}
	}
	return experimental
}
