package extensions

import (
	"encoding/json"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/pkg/errors"
)

const (
	DependenciesKey    = "dependencies"
	DependenciesSchema = "https://cnab.io/specs/v1/dependencies.schema.json"
)

// Dependencies describes the set of custom extension metadata associated with the dependencies spec
// https://github.com/deislabs/cnab-spec/blob/master/500-CNAB-dependencies.md
type Dependencies struct {
	// Requires is a list of bundles required by this bundle
	Requires []Dependency `json:"requires,omitempty" mapstructure:"requires"`
}

// Dependency describes a dependency on another bundle
type Dependency struct {
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

func LoadDependencies(bun *bundle.Bundle) (*Dependencies, error) {
	data := bun.Custom[DependenciesKey]

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal the untyped dependencies extension data %q", string(dataB))
	}

	deps := &Dependencies{}
	err = json.Unmarshal(dataB, deps)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal the dependencies extension %q", string(dataB))
	}

	return deps, nil
}
