package secrets

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Set is an actual set of resolved values.
// This is the output of resolving a parameter or credential set file.
type Set map[string]string

// Merge merges a second Set into the base.
//
// Duplicate names are not allow and will result in an
// error, this is the case even if the values are identical.
func (s Set) Merge(s2 Set) error {
	for k, v := range s2 {
		if _, ok := s[k]; ok {
			return fmt.Errorf("ambiguous value resolution: %q is already present in base sets, cannot merge", k)
		}
		s[k] = v
	}
	return nil
}

// Strategy represents a strategy for determining the value of a parameter or credential
type Strategy struct {
	// Name is the name of the parameter or credential.
	Name string `json:"name" yaml:"name"`
	// Source is the location of the value.
	// During resolution, the source will be loaded, and the result temporarily placed
	// into Value.
	Source Source `json:"source,omitempty" yaml:"source,omitempty"`
	// Value holds the parameter or credential value.
	// When a parameter or credential is loaded, it is loaded into this field. In all
	// other cases, it is empty. This field is omitted during serialization.
	Value string `json:"-" yaml:"-"`
}

// Source represents a strategy for loading a value from local host.
type Source struct {
	Key   string
	Value string
}

func (s Source) MarshalRaw() interface{} {
	if s.Key == "" {
		return nil
	}
	return map[string]interface{}{s.Key: s.Value}
}

func (s *Source) UnmarshalRaw(raw map[string]interface{}) error {
	switch len(raw) {
	case 0:
		s.Key = ""
		s.Value = ""
		return nil
	case 1:
		for k, v := range raw {
			s.Key = k
			if value, ok := v.(string); ok {
				s.Value = value
			} else {
				s.Value = fmt.Sprintf("%v", s.Value)
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

type StrategyList []Strategy

func (l StrategyList) Less(i, j int) bool {
	return l[i].Name < l[j].Name
}

func (l StrategyList) Swap(i, j int) {
	tmp := l[i]
	l[i] = l[j]
	l[j] = tmp
}

func (l StrategyList) Len() int {
	return len(l)
}
