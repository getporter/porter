package experimental

const (
	// NoopFeature is a placeholder feature flag that allows us to test our feature flag functions even when there are no active feature flags
	NoopFeature = "no-op"

	// DependenciesV2 is the name of the experimental feature flag for PEP003 - Advanced Dependencies.
	DependenciesV2 = "dependencies-v2"
)

// FeatureFlags is an enum of possible feature flags
type FeatureFlags int

const (
	// FlagNoopFeature is a placeholder feature flag that allows us to test our feature flag functions even when there are no active feature flags
	FlagNoopFeature FeatureFlags = iota + 1

	// FlagDependenciesV2 gates the changes from PEP003 - Advanced Dependencies.
	FlagDependenciesV2
)

// ParseFlags converts a list of feature flag names into a bit map for faster lookups.
func ParseFlags(flags []string) FeatureFlags {
	var experimental FeatureFlags
	for _, flag := range flags {
		switch flag {
		case NoopFeature:
			experimental = experimental | FlagNoopFeature
		case DependenciesV2:
			experimental = experimental | FlagDependenciesV2
		}
	}
	return experimental
}
