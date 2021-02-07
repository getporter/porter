package experimental

const (
	// BuildDrivers experimental flag
	BuildDrivers = "build-drivers"
)

type FeatureFlags int

const (
	// FlagBuildDrivers indicates if configurable build drivers are enabled.
	FlagBuildDrivers FeatureFlags = iota + 1
)

func ParseFlags(flags []string) FeatureFlags {
	var experimental FeatureFlags
	for _, flag := range flags {
		switch flag {
		case BuildDrivers:
			experimental = experimental | FlagBuildDrivers
		}
	}
	return experimental
}
