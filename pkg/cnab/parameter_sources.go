package cnab

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	// ParameterSourcesExtensionShortHand is the short suffix of the ParameterSourcesExtensionKey.
	ParameterSourcesExtensionShortHand = "parameter-sources"

	// ParameterSourcesExtensionKey represents the full key for the Parameter Sources Extension.
	ParameterSourcesExtensionKey = OfficialExtensionsPrefix + ParameterSourcesExtensionShortHand

	// ParameterSourcesExtensionSchema represents the schema for the Docker Extension.
	ParameterSourcesSchema = "https://cnab.io/v1/parameter-sources.schema.json"
	// ParameterSourceTypeOutput defines a type of parameter source that is provided by a bundle output.
	ParameterSourceTypeOutput = "output"
	// ParameterSourceTypeDependencyOutput defines a type of parameter source that is provided by a bundle's dependency
	// output.
	ParameterSourceTypeDependencyOutput = "dependencies.output"
)

// ParameterSourcesExtension represents a required extension that specifies how
// to default parameter values.
var ParameterSourcesExtension = RequiredExtension{
	Shorthand: ParameterSourcesExtensionShortHand,
	Key:       ParameterSourcesExtensionKey,
	Schema:    ParameterSourcesSchema,
	Reader:    ParameterSourcesReader,
}

// ParameterSources describes the set of custom extension metadata associated
// with the Parameter Sources extension
type ParameterSources map[string]ParameterSource

// SetParameterFromOutput creates an entry in the parameter sources section setting
// the parameter's value using the specified output's value.
func (ps *ParameterSources) SetParameterFromOutput(parameter string, output string) {
	if *ps == nil {
		*ps = ParameterSources{}
	}

	(*ps)[parameter] = ParameterSource{
		Priority: []string{ParameterSourceTypeOutput},
		Sources: ParameterSourceMap{
			ParameterSourceTypeOutput: OutputParameterSource{OutputName: output},
		},
	}
}

// SetParameterFromDependencyOutput creates an entry in the parameter sources section setting
// the parameter's value using the specified dependency's output value.
func (ps *ParameterSources) SetParameterFromDependencyOutput(parameter string, dep string, output string) {
	if *ps == nil {
		*ps = ParameterSources{}
	}

	(*ps)[parameter] = ParameterSource{
		Priority: []string{ParameterSourceTypeDependencyOutput},
		Sources: ParameterSourceMap{
			ParameterSourceTypeDependencyOutput: DependencyOutputParameterSource{
				Dependency: dep,
				OutputName: output},
		},
	}
}

type ParameterSource struct {
	// Priority is an array of source types in the priority order that they should be used to
	// populated the parameter.
	Priority []string `json:"priority" mapstructure:"priority"`

	// Sources is a map of key/value pairs of a source type and definition for
	// the parameter value.
	Sources ParameterSourceMap `json:"sources" mapstructure:"sources"`
}

// ListSourcesByPriority returns the parameter sources by the requested priority,
// if none is specified, they are unsorted.
func (s ParameterSource) ListSourcesByPriority() []ParameterSourceDefinition {
	sources := make([]ParameterSourceDefinition, 0, len(s.Sources))
	if len(s.Priority) == 0 {
		for _, source := range s.Sources {
			sources = append(sources, source)
		}
	} else {
		for _, sourceType := range s.Priority {
			sources = append(sources, s.Sources[sourceType])
		}
	}
	return sources
}

type ParameterSourceMap map[string]ParameterSourceDefinition

func (m *ParameterSourceMap) UnmarshalJSON(data []byte) error {
	if *m == nil {
		*m = ParameterSourceMap{}
	}

	var rawMap map[string]interface{}
	err := json.Unmarshal(data, &rawMap)
	if err != nil {
		return err
	}

	for sourceKey, sourceDef := range rawMap {
		rawDef, err := json.Marshal(sourceDef)
		if err != nil {
			return fmt.Errorf("error re-marshaling parameter source definition: %w", err)
		}

		switch sourceKey {
		case ParameterSourceTypeOutput:
			var output OutputParameterSource
			err := json.Unmarshal(rawDef, &output)
			if err != nil {
				return fmt.Errorf("invalid parameter source definition for key %s: %w", sourceKey, err)
			}
			(*m)[ParameterSourceTypeOutput] = output
		case ParameterSourceTypeDependencyOutput:
			var depOutput DependencyOutputParameterSource
			err := json.Unmarshal(rawDef, &depOutput)
			if err != nil {
				return fmt.Errorf("invalid parameter source definition for key %s: %w", sourceKey, err)
			}
			(*m)[ParameterSourceTypeDependencyOutput] = depOutput
		default:
			return fmt.Errorf("unsupported parameter source key %s", sourceKey)
		}
	}

	return nil
}

type ParameterSourceDefinition interface {
}

// OutputParameterSource represents a parameter that is set using the value from
// a bundle output.
type OutputParameterSource struct {
	OutputName string `json:"name" mapstructure:"name"`
}

// DependencyOutputParameterSource represents a parameter that is set using the value
// from a bundle's dependency output.
type DependencyOutputParameterSource struct {
	Dependency string `json:"dependency" mapstructure:"dependency"`
	OutputName string `json:"name" mapstructure:"name"`
}

// ReadParameterSources is a convenience method for returning a bonafide
// ParameterSources reference after reading from the applicable section from
// the provided bundle
func (b ExtendedBundle) ReadParameterSources() (ParameterSources, error) {
	raw, err := b.ParameterSourcesReader()
	if err != nil {
		return nil, err
	}

	ps, ok := raw.(ParameterSources)
	if !ok {
		return nil, errors.New("unable to read parameter sources extension data")
	}

	return ps, nil
}

// ParameterSourcesReader is a Reader for the ParameterSourcesExtension,
// which reads from the applicable section in the provided bundle and
// returns the raw data in the form of an interface
func ParameterSourcesReader(bun ExtendedBundle) (interface{}, error) {
	return bun.ParameterSourcesReader()
}

// ParameterSourcesReader is a Reader for the ParameterSourcesExtension,
// which reads from the applicable section in the provided bundle and
// returns the raw data in the form of an interface
func (b ExtendedBundle) ParameterSourcesReader() (interface{}, error) {
	data, ok := b.Custom[ParameterSourcesExtensionKey]
	if !ok {
		return nil, errors.New("attempted to read parameter sources from bundle but none are defined")
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the untyped %q extension data %q: %w",
			ParameterSourcesExtensionKey, string(dataB), err)
	}

	ps := ParameterSources{}
	err = json.Unmarshal(dataB, &ps)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal the %q extension %q: %w",
			ParameterSourcesExtensionKey, string(dataB), err)
	}

	return ps, nil
}

// SupportsParameterSources checks if the bundle supports parameter sources.
func (b ExtendedBundle) SupportsParameterSources() bool {
	return b.SupportsExtension(ParameterSourcesExtensionKey)
}

// GetParameterSources checks if the parameter sources extension is present and returns its
// extension configuration.
func (e ProcessedExtensions) GetParameterSources() (ParameterSources, bool, error) {
	rawExt, required := e[ParameterSourcesExtensionKey]

	ext, ok := rawExt.(ParameterSources)
	if !ok && required {
		return ParameterSources{}, required, fmt.Errorf("unable to parse Parameter Sources extension config: %+v", rawExt)
	}

	return ext, required, nil
}

// HasParameterSources returns whether or not the bundle has parameter sources defined.
func (b ExtendedBundle) HasParameterSources() bool {
	_, ok := b.Custom[ParameterSourcesExtensionKey]
	return ok
}

// ParameterHasSource determines if the specified parameter has a parameter
// source defined.
func (b ExtendedBundle) ParameterHasSource(paramName string) bool {
	sources, err := b.ReadParameterSources()
	if err != nil {
		return false
	}
	_, hasSource := sources[paramName]
	return hasSource
}
