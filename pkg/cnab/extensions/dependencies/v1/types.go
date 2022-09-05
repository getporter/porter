package v1

// Dependencies describes the set of custom extension metadata associated with the dependencies spec
// https://github.com/cnabio/cnab-spec/blob/master/500-CNAB-dependencies.md
type Dependencies struct {
	// Sequence is a list to order the dependencies
	Sequence []string `json:"sequence,omitempty" mapstructure:"sequence"`

	// Requires is a list of bundles required by this bundle
	Requires map[string]Dependency `json:"requires,omitempty" mapstructure:"requires"`
}

// ListBySequence returns the dependencies by the defined sequence,
// if none is specified, they are unsorted.
func (d Dependencies) ListBySequence() []Dependency {
	deps := make([]Dependency, 0, len(d.Requires))
	if len(d.Sequence) > 0 && len(d.Sequence) == len(d.Requires) {
		for _, depName := range d.Sequence {
			dep := d.Requires[depName]
			dep.Name = depName
			deps = append(deps, dep)
		}
	} else {
		for depName, dep := range d.Requires {
			dep.Name = depName
			deps = append(deps, dep)
		}
	}
	return deps
}

// Dependency describes a dependency on another bundle
type Dependency struct {
	// Name of the dependency
	Name string `json:"name" mapstructure:"name"`

	// Bundle is the location of the bundle in a registry, for example REGISTRY/NAME:TAG
	Bundle string `json:"bundle" mapstructure:"bundle"`

	// Version is a set of allowed versions
	Version *DependencyVersion `json:"version,omitempty" mapstructure:"version"`
}

// DependencyVersion is a set of allowed versions for a dependency
type DependencyVersion struct {
	// Ranges of semantic versions, with or without the leading v prefix, allowed by the dependency
	Ranges []string `json:"ranges,omitempty" mapstructure:"ranges"`

	// AllowPrereleases specifies if prerelease versions can satisfy the dependency
	AllowPrereleases bool `json:"prereleases" mapstructure:"prereleases"`
}
