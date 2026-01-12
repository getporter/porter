package encoding

import (
	"encoding/json"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// MapElement is the in-memory representation of the item when stored in a map.
type MapElement interface {
	// ToArrayEntry converts to the representation of the item when stored in an
	// array.
	ToArrayEntry(key string) ArrayElement
}

// ArrayElement is the representation of the item when stored in an array, and
// includes the key under which the element was stored in the original map.
type ArrayElement interface {
	// GetKey returns the unique item key.
	GetKey() string

	// ToMapEntry converts to the representation of the item when stored in a map.
	ToMapEntry() MapElement
}

// ArrayEncodedMap is a map that is represented as an array when marshaled to json/yaml.
// MapElement is the type of the elements when stored in a map and ArrayElement is the type of the elements when stored in an array.
type ArrayEncodedMap[T MapElement, K ArrayElement] struct {
	items map[string]T
}

// NewArrayEncodedMap initializes an empty ArrayEncodedMap.
func NewArrayEncodedMap[T MapElement, K ArrayElement]() ArrayEncodedMap[T, K] {
	return MakeArrayEncodedMap[T, K](0)
}

// MakeArrayEncodedMap allocates memory for the specified number of elements.
func MakeArrayEncodedMap[T MapElement, K ArrayElement](len int) ArrayEncodedMap[T, K] {
	return ArrayEncodedMap[T, K]{
		items: make(map[string]T, len),
	}
}

// Len returns the number of items.
func (m *ArrayEncodedMap[T, K]) Len() int {
	if m == nil {
		return 0
	}
	return len(m.items)
}

// Items returns a copy of the items, and is intended for use with the range
// operator.
// Use ItemsUnsafe() to directly manipulate the backing items map.
func (m *ArrayEncodedMap[T, K]) Items() map[string]T {
	if m == nil {
		return nil
	}

	result := make(map[string]T, len(m.items))
	for k, v := range m.items {
		result[k] = v
	}
	return result
}

// ItemsSorted returns a copy of the items, in a sorted array, and is intended
// for using with serialization and consistently ranging over the items, in tests
// or printing output to the console.
func (m *ArrayEncodedMap[T, K]) ItemsSorted() []K {
	if m == nil {
		return nil
	}

	result := make([]K, len(m.items))
	i := 0
	for k, v := range m.items {
		// I can't figure out how to constrain T such that ToArrayEntry returns K, so I'm doing a cast
		result[i] = v.ToArrayEntry(k).(K)
		i++
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].GetKey() < result[j].GetKey()
	})

	return result
}

// ItemsUnsafe returns the backing items map. It is intended for optimizing
// memory usage when iterating over the items to convert them into alternative
// representations.
func (m *ArrayEncodedMap[T, K]) ItemsUnsafe() *map[string]T {
	if m == nil {
		return nil
	}

	if m.items == nil {
		m.items = make(map[string]T)
	}

	return &m.items
}

// Get returns the specified element by its key.
func (m *ArrayEncodedMap[T, K]) Get(key string) (T, bool) {
	if m == nil {
		return *new(T), false
	}

	entry, ok := m.items[key]
	return entry, ok
}

// Set the specified element by its key, overwriting previous values.
func (m *ArrayEncodedMap[T, K]) Set(key string, entry T) {
	if m.items == nil {
		m.items = make(map[string]T, 1)
	}

	m.items[key] = entry
}

// Remove the specified element by its key.
func (m *ArrayEncodedMap[T, K]) Remove(key string) {
	if m == nil {
		return
	}
	delete(m.items, key)
}

// MarshalRaw is the common Marshal implementation between YAML and JSON.
func (m *ArrayEncodedMap[T, K]) MarshalRaw() interface{} {
	if m == nil {
		return nil
	}

	var raw []ArrayElement
	if m.items == nil {
		return raw
	}

	raw = make([]ArrayElement, 0, len(m.items))
	for k, v := range m.items {
		raw = append(raw, v.ToArrayEntry(k))
	}
	sort.SliceStable(raw, func(i, j int) bool {
		return raw[i].GetKey() < raw[j].GetKey()
	})
	return raw
}

// UnmarshalRaw is the common Marshal implementation between YAML and JSON.
func (m *ArrayEncodedMap[T, K]) UnmarshalRaw(raw []K) error {
	if m == nil {
		*m = ArrayEncodedMap[T, K]{}
	}

	m.items = make(map[string]T, len(raw))
	for _, rawItem := range raw {
		if _, hasKey := m.items[rawItem.GetKey()]; hasKey {
			return fmt.Errorf("cannot unmarshal source map: duplicate key found '%s'", rawItem.GetKey())
		}
		item := rawItem.ToMapEntry()
		typedItem, ok := item.(T)
		if !ok {
			return fmt.Errorf("invalid ArrayEncodedMap generic types, ArrayElement %T returned a %T from ToMapEntry(), when it should return %T", rawItem, item, *new(T))
		}
		m.items[rawItem.GetKey()] = typedItem
	}
	return nil
}

// MarshalJSON marshals the items to JSON.
func (m *ArrayEncodedMap[T, K]) MarshalJSON() ([]byte, error) {
	raw := m.MarshalRaw()
	return json.Marshal(raw)
}

// UnmarshalJSON unmarshals the items in the specified JSON.
func (m *ArrayEncodedMap[T, K]) UnmarshalJSON(data []byte) error {
	var raw []K
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	return m.UnmarshalRaw(raw)
}

// MarshalYAML marshals the items to YAML.
func (m *ArrayEncodedMap[T, K]) MarshalYAML() (interface{}, error) {
	if m == nil {
		return nil, nil
	}
	return m.MarshalRaw(), nil
}

// UnmarshalYAML unmarshals the items in the specified YAML.
func (m *ArrayEncodedMap[T, K]) UnmarshalYAML(value *yaml.Node) error {
	var raw []K
	if err := value.Decode(&raw); err != nil {
		return err
	}
	return m.UnmarshalRaw(raw)
}

// Merge applies the specified values on top of a base set of values. When a
// key exists in both sets, use the value from the overrides.
func (m *ArrayEncodedMap[T, K]) Merge(overrides *ArrayEncodedMap[T, K]) *ArrayEncodedMap[T, K] {
	result := make(map[string]T, m.Len())
	if m != nil {
		for k, v := range m.items {
			result[k] = v
		}
	}

	if overrides != nil {
		// If the name is in the base, overwrite its value with the override provided
		for k, v := range overrides.items {
			result[k] = v
		}
	}

	return &ArrayEncodedMap[T, K]{items: result}
}
