package secrets

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cnabio/cnab-go/valuesource"
	"gopkg.in/yaml.v3"
)

// Set is an actual set of resolved values.
// This is the output of resolving a parameter or credential set file.
type Set map[string]string

// IsValid determines if the provided key (designating a name of a parameter
// or credential) is included in the provided set
func (s Set) IsValid(key string) bool {
	for name := range s {
		if name == key {
			return true
		}
	}
	return false
}

// ToCNAB converts this to a type accepted by the cnab-go runtime.
func (s Set) ToCNAB() valuesource.Set {
	return valuesource.Set(s)
}

// SourceMap maps from a parameter or credential name to a source strategy for resolving its value.
type SourceMap struct {
	// Name is the name of the parameter or credential.
	Name string `json:"name" yaml:"name"`

	// Source defines a strategy for resolving a value from the specified source.
	Source Source `json:"source,omitempty" yaml:"source,omitempty"`

	// ResolvedValue holds the resolved parameter or credential value.
	// When a parameter or credential is resolved, it is loaded into this field. In all
	// other cases, it is empty. This field is omitted during serialization.
	ResolvedValue string `json:"-" yaml:"-"`
}

// Source specifies how to resolve a parameter or credential from an external
// source.
type Source struct {
	// Strategy to resolve the source value, e.g. "secret" or "env".
	Strategy string

	// Hint to the strategy handler on how to resolve the value.
	// For example the name of the secret in a secret store or name of an environment variable.
	Hint string
}

func (s Source) MarshalRaw() interface{} {
	if s.Strategy == "" {
		return nil
	}
	return map[string]interface{}{s.Strategy: s.Hint}
}

func (s *Source) UnmarshalRaw(raw map[string]interface{}) error {
	switch len(raw) {
	case 0:
		s.Strategy = ""
		s.Hint = ""
		return nil
	case 1:
		for k, v := range raw {
			s.Strategy = k
			if value, ok := v.(string); ok {
				s.Hint = value
			} else {
				s.Hint = fmt.Sprintf("%v", s.Hint)
			}
		}
		return nil
	default:
		return errors.New("multiple key/value pairs specified for source but only one may be defined")
	}
}

var (
	_ json.Marshaler   = Source{}
	_ json.Unmarshaler = &Source{}
	_ yaml.Marshaler   = Source{}
	_ yaml.Unmarshaler = &Source{}
)

func (s Source) MarshalJSON() ([]byte, error) {
	raw := s.MarshalRaw()
	return json.Marshal(raw)
}

func (s *Source) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	return s.UnmarshalRaw(raw)
}

func (s *Source) UnmarshalYAML(value *yaml.Node) error {
	var raw map[string]interface{}
	err := value.Decode(&raw)
	if err != nil {
		return err
	}
	return s.UnmarshalRaw(raw)
}

func (s Source) MarshalYAML() (interface{}, error) {
	return s.MarshalRaw(), nil
}

// SourceMapList is a list of mappings that can be access via index or the item name.
type SourceMapList []SourceMap

func (l SourceMapList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l SourceMapList) Swap(i, j int) {
	tmp := l[i]
	l[i] = l[j]
	l[j] = tmp
}

func (l SourceMapList) Len() int {
	return len(l)
}

// HasName determines if the specified name is defined in the set.
func (l SourceMapList) HasName(name string) bool {
	_, ok := l.GetByName(name)
	return ok
}

// GetByName returns the resolution strategy for the specified name and a bool indicating if it was found.
func (l SourceMapList) GetByName(name string) (SourceMap, bool) {
	for _, item := range l {
		if item.Name == name {
			return item, true
		}
	}

	return SourceMap{}, false
}

// GetResolvedValue returns the resolved value of the specified name and a bool indicating if it was found.
// You must resolve the value before calling, it does not do resolution for you.
func (l SourceMapList) GetResolvedValue(name string) (interface{}, bool) {
	item, ok := l.GetByName(name)
	if ok {
		return item.ResolvedValue, true
	}

	return nil, false
}

// ToResolvedValues converts the items to a map of key/value pairs, with the resolved values represented as CNAB-compatible strings
func (l SourceMapList) ToResolvedValues() map[string]string {
	values := make(map[string]string, len(l))
	for _, item := range l {
		values[item.Name] = item.ResolvedValue
	}
	return values
}

// Merge applies the specified values on top of a base set of values. When a
// name exists in both sets, use the value from the overrides
func (l SourceMapList) Merge(overrides SourceMapList) SourceMapList {
	result := l

	// Make a lookup from the name to its index in result so that we can quickly find a
	// named item while merging
	lookup := make(map[string]int, len(result))
	for i, item := range result {
		lookup[item.Name] = i
	}

	for _, item := range overrides {
		// If the name is in the base, overwrite its value with the override provided
		if i, ok := lookup[item.Name]; ok {
			result[i].Source = item.Source

			// Just in case the value was resolved, include in the merge results
			result[i].ResolvedValue = item.ResolvedValue
		} else {
			// Append the override to the list of results if it's not in the base set of values
			result = append(result, item)
		}
	}

	return result
}
