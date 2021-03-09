package yaml

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

func Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

func Marshal(value interface{}) ([]byte, error) {
	b := bytes.Buffer{}
	encoder := yaml.NewEncoder(&b)
	encoder.SetIndent(2) // this is what you're looking for
	err := encoder.Encode(value)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
