package cnab

import (
	"encoding/json"
	"errors"
	"fmt"

	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
)

const (
	// DependenciesV2ExtensionShortHand is the short suffix of the DependenciesV2ExtensionKey
	DependenciesV2ExtensionShortHand = "dependencies@v2"

	// DependenciesV2ExtensionKey represents the full key for the DependenciesV2Extension.
	DependenciesV2ExtensionKey = "org.getporter." + DependenciesV2ExtensionShortHand

	// DependenciesV2Schema represents the schema for the DependenciesV2 Extension
	DependenciesV2Schema = "https://porter.sh/extensions/dependencies/v2/schema.json"
)

// DependenciesV2Extension represents the required extension to enable dependencies
var DependenciesV2Extension = RequiredExtension{
	Shorthand: DependenciesV2ExtensionShortHand,
	Key:       DependenciesV2ExtensionKey,
	Schema:    DependenciesV2Schema,
	Reader: func(b ExtendedBundle) (interface{}, error) {
		return b.DependencyV2Reader()
	},
}

// ReadDependenciesV2 is a convenience method for returning a bonafide
// DependenciesV2 reference after reading from the applicable section from
// the provided bundle
func (b ExtendedBundle) ReadDependenciesV2() (v2.Dependencies, error) {
	raw, err := b.DependencyV2Reader()
	if err != nil {
		return v2.Dependencies{}, err
	}

	deps, ok := raw.(v2.Dependencies)
	if !ok {
		return v2.Dependencies{}, errors.New("unable to read dependencies v2 extension data")
	}

	// Return the dependencies
	return deps, nil
}

// DependencyV2Reader is a Reader for the DependenciesV2Extension, which reads
// from the applicable section in the provided bundle and returns the raw
// data in the form of an interface
func (b ExtendedBundle) DependencyV2Reader() (interface{}, error) {
	data, ok := b.Custom[DependenciesV2ExtensionKey]
	if !ok {
		return nil, fmt.Errorf("attempted to read %s extension data from bundle but none are defined", DependenciesV2ExtensionKey)
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the untyped %s extension data %q: %w", DependenciesV2ExtensionKey, string(dataB), err)
	}

	//Note: For dependency.Name to be set properly ReadDependencyV2
	// *must* be called.
	//todo: make it so that ReadDependencyV2 is only able to be exported.
	deps := v2.Dependencies{}
	err = json.Unmarshal(dataB, &deps)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal the %s extension %q: %w", DependenciesV2ExtensionKey, string(dataB), err)
	}

	return deps, nil
}

// SupportsDependenciesV2 checks if the bundle supports dependencies
func (b ExtendedBundle) SupportsDependenciesV2() bool {
	return b.SupportsExtension(DependenciesV2ExtensionKey)
}

// HasDependenciesV2 returns whether the bundle has v2 dependencies defined.
func (b ExtendedBundle) HasDependenciesV2() bool {
	_, ok := b.Custom[DependenciesV2ExtensionKey]
	return ok
}
