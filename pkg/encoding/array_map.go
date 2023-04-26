package encoding

import (
	"encoding/json"
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// MapElement is the in-memory representation of the item when stored in a map.
type MapElement interface {
	ToArrayEntry(key string) ArrayElement
}

// ArrayElement is the representation of the item when encoded to an array in yaml or json.
type ArrayElement interface {
	GetKey() string
	ToMapEntry() MapElement
}

// ArrayMap is a map that is represented as an array when marshaled.
type ArrayMap[T MapElement, K ArrayElement] struct {
	items map[string]T
}

// TODO(carolyn): Can I make this work without K?
func MakeArrayMap[T MapElement, K ArrayElement](len int) ArrayMap[T, K] {
	return ArrayMap[T, K]{
		items: make(map[string]T, len),
	}
}

func (m *ArrayMap[T, K]) Len() int {
	if m == nil {
		return 0
	}
	return len(m.items)
}

func (m *ArrayMap[T, K]) Items() map[string]T {
	if m == nil {
		return nil
	}

	result := make(map[string]T, len(m.items))
	for k, v := range m.items {
		result[k] = v
	}
	return result
}

func (m *ArrayMap[T, K]) ItemsSorted() []K {
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

func (m *ArrayMap[T, K]) ItemsUnsafe() map[string]T {
	if m == nil {
		return nil
	}

	if m.items == nil {
		m.items = make(map[string]T)
	}

	return m.items
}

func (m *ArrayMap[T, K]) Get(key string) (T, bool) {
	if m == nil {
		return *new(T), false
	}

	entry, ok := m.items[key]
	return entry, ok
}

func (m *ArrayMap[T, K]) Set(key string, entry T) {
	if m.items == nil {
		m.items = make(map[string]T, 1)
	}

	m.items[key] = entry
}

func (m *ArrayMap[T, K]) Remove(key string) {
	if m == nil {
		return
	}
	delete(m.items, key)
}

func (m *ArrayMap[T, K]) MarshalRaw() interface{} {
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

func (m *ArrayMap[T, K]) UnmarshalRaw(raw []K) error {
	// If there's nothing to import, stop early and allow the map to keep its original value
	// So if someone unmarshalled into a nil map, it stays nil.
	// This more closely matches how the stdlib encoders work
	if len(raw) == 0 {
		return nil
	}

	m.items = make(map[string]T, len(raw))
	for _, rawItem := range raw {
		if _, hasKey := m.items[rawItem.GetKey()]; hasKey {
			return fmt.Errorf("cannot unmarshal source map: duplicate key found '%s'", rawItem.GetKey())
		}
		item := rawItem.ToMapEntry()
		typedItem, ok := item.(T)
		if !ok {
			return fmt.Errorf("invalid ArrayMap generic types, ArrayElement %T returned a %T from ToMapEntry(), when it should return %T", rawItem, item, *new(T))
		}
		m.items[rawItem.GetKey()] = typedItem
	}
	return nil
}

func (m *ArrayMap[T, K]) MarshalJSON() ([]byte, error) {
	raw := m.MarshalRaw()
	return json.Marshal(raw)
}

func (m *ArrayMap[T, K]) UnmarshalJSON(data []byte) error {
	var raw []K
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	return m.UnmarshalRaw(raw)
}

func (m *ArrayMap[T, K]) UnmarshalYAML(value *yaml.Node) error {
	var raw []K
	if err := value.Decode(&raw); err != nil {
		return err
	}
	return m.UnmarshalRaw(raw)
}

func (m *ArrayMap[T, K]) MarshalYAML() (interface{}, error) {
	if m == nil {
		return nil, nil
	}
	return m.MarshalRaw(), nil
}

// Merge applies the specified values on top of a base set of values. When a
// name exists in both sets, use the value from the overrides
func (m *ArrayMap[T, K]) Merge(overrides *ArrayMap[T, K]) *ArrayMap[T, K] {
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

	return &ArrayMap[T, K]{items: result}
}
