package secrets

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cnabio/cnab-go/secrets/host"
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

// HardCodedValue generates a hard-coded value source mapping that contains a resolved value.
func HardCodedValue(value string) ValueMapping {
	return ValueMapping{
		Source: Source{
			Strategy: host.SourceValue,
			Hint:     value},
		ResolvedValue: value}
}

// HardCodedValueStrategy generates a hard-coded value strategy.
// TODO(carolyn): Remove name arg
func HardCodedValueStrategy(value string) Source {
	return Source{
		Strategy: host.SourceValue,
		Hint:     value}
}
