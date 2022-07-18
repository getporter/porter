package cnab

import (
	"encoding/json"
	"errors"
	"fmt"

	depsv1 "get.porter.sh/porter/pkg/cnab/dependencies/v1"
)

const (
	// DependenciesV1ExtensionShortHand is the short suffix of the DependenciesV1ExtensionKey
	DependenciesV1ExtensionShortHand = "dependencies"

	// DependenciesV1ExtensionKey represents the full key for the DependenciesV1Extension.
	DependenciesV1ExtensionKey = OfficialExtensionsPrefix + DependenciesV1ExtensionShortHand

	// DependenciesV1Schema represents the schema for the Dependencies Extension
	DependenciesV1Schema = "https://cnab.io/v1/dependencies.schema.json"
)

// DependenciesV1Extension represents the required extension to enable dependencies
var DependenciesV1Extension = RequiredExtension{
	Shorthand: DependenciesV1ExtensionShortHand,
	Key:       DependenciesV1ExtensionKey,
	Schema:    DependenciesV1Schema,
	Reader: func(b ExtendedBundle) (interface{}, error) {
		return b.DependencyV1Reader()
	},
}

// ReadDependenciesV1 is a convenience method for returning a bonafide
// Dependencies reference after reading from the applicable section from
// the provided bundle
func (b ExtendedBundle) ReadDependenciesV1() (depsv1.Dependencies, error) {
	raw, err := b.DependencyV1Reader()
	if err != nil {
		return depsv1.Dependencies{}, err
	}

	deps, ok := raw.(depsv1.Dependencies)
	if !ok {
		return depsv1.Dependencies{}, errors.New("unable to read dependencies extension data")
	}

	// Return the dependencies
	return deps, nil
}

// DependencyV1Reader is a Reader for the DependenciesV1Extension, which reads
// from the applicable section in the provided bundle and returns the raw
// data in the form of an interface
func (b ExtendedBundle) DependencyV1Reader() (interface{}, error) {
	data, ok := b.Custom[DependenciesV1ExtensionKey]
	if !ok {
		return nil, fmt.Errorf("attempted to read dependencies from bundle but none are defined")
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the untyped dependencies extension data %q: %w", string(dataB), err)
	}

	deps := depsv1.Dependencies{}
	err = json.Unmarshal(dataB, &deps)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal the dependencies extension %q: %w", string(dataB), err)
	}

	return deps, nil
}

// SupportsDependenciesV1 checks if the bundle supports dependencies
func (b ExtendedBundle) SupportsDependenciesV1() bool {
	return b.SupportsExtension(DependenciesV1ExtensionKey)
}

// HasDependenciesV1 returns whether the bundle has parameter sources defined.
func (b ExtendedBundle) HasDependenciesV1() bool {
	_, ok := b.Custom[DependenciesV1ExtensionKey]
	return ok
}
