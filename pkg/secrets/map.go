package secrets

import (
	"encoding/json"

	"get.porter.sh/porter/pkg/encoding"
	"gopkg.in/yaml.v3"
)

type Map struct {
	*encoding.ArrayMap[ValueMapping, NamedValueMapping]
}

func NewMap() Map {
	return Map{
		ArrayMap: &encoding.ArrayMap[ValueMapping, NamedValueMapping]{},
	}
}

func MakeMap(len int) Map {
	items := encoding.MakeArrayMap[ValueMapping, NamedValueMapping](len)
	return Map{
		ArrayMap: &items,
	}
}

func (m Map) Merge(overrides Map) Map {
	result := m.ArrayMap.Merge(overrides.ArrayMap)
	return Map{ArrayMap: result}
}

func (m Map) ToResolvedValues() map[string]string {
	values := make(map[string]string, m.Len())
	for k, v := range m.ItemsUnsafe() {
		values[k] = v.ResolvedValue
	}
	return values
}

func (m *Map) UnmarshalJSON(b []byte) error {
	// Ensure that ArrayMap is not nil before its custom UnmarshalJson is called
	// Otherwise it causes a panic
	if m.ArrayMap == nil {
		m.ArrayMap = &encoding.ArrayMap[ValueMapping, NamedValueMapping]{}
	}

	// Cheap trick to unmarshal without re-triggering this UnmarshalJson call
	type RawMap Map
	return json.Unmarshal(b, (*RawMap)(m))
}

func (m *Map) UnmarshalYAML(value *yaml.Node) error {
	// Ensure that ArrayMap is not nil before its custom UnmarshalYAML is called
	// Otherwise it causes a panic
	if m.ArrayMap == nil {
		m.ArrayMap = &encoding.ArrayMap[ValueMapping, NamedValueMapping]{}
	}

	// Cheap trick to unmarshal without re-triggering this UnmarshalYAML call
	type RawMap Map
	return value.Decode((*RawMap)(m))
}

var _ encoding.MapElement = ValueMapping{}

// ValueMapping maps from a parameter or credential name to a source strategy for resolving its value.
type ValueMapping struct {
	// Source defines a strategy for resolving a value from the specified source.
	Source Source `json:"source,omitempty" yaml:"source,omitempty"`

	// ResolvedValue holds the resolved parameter or credential value.
	// When a parameter or credential is resolved, it is loaded into this field. In all
	// other cases, it is empty. This field is omitted during serialization.
	ResolvedValue string `json:"-" yaml:"-"`
}

func (v ValueMapping) ToArrayEntry(key string) encoding.ArrayElement {
	return NamedValueMapping{
		Name:   key,
		Source: v.Source,
	}
}

// NamedValueMapping is the representation of a ValueMapping in an array,
// We use it to unmarshal the yaml, and then convert it into our internal representation
// where the ValueMapping is in a Go map instead of an array.
type NamedValueMapping struct {
	// Name is the name of the parameter or credential.
	Name string `json:"name" yaml:"name"`

	// Source defines a strategy for resolving a value from the specified source.
	Source Source `json:"source" yaml:"source"`

	// ResolvedValue holds the resolved parameter or credential value.
	// When a parameter or credential is resolved, it is loaded into this field. In all
	// other cases, it is empty. This field is omitted during serialization.
	ResolvedValue string `json:"-" yaml:"-"`
}

func (r NamedValueMapping) ToValueMapping() ValueMapping {
	return ValueMapping{
		Source:        r.Source,
		ResolvedValue: r.ResolvedValue,
	}
}

func (r NamedValueMapping) ToMapEntry() encoding.MapElement {
	return r.ToValueMapping()
}

func (r NamedValueMapping) GetKey() string {
	return r.Name
}
