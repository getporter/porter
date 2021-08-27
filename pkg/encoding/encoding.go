package encoding

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/carolynvs/aferox"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	Yaml = "yaml"
	Json = "json"
	Toml = "toml"
)

// MarshalFile encodes the specified struct to a file.
// Supported file extensions are: yaml, yml, json, and toml.
func MarshalFile(fs aferox.Aferox, path string, in interface{}) error {
	format := strings.TrimPrefix(filepath.Ext(path), ".")
	data, err := Marshal(format, in)
	if err != nil {
		return err
	}
	return fs.WriteFile(path, data, 0700)
}

// MarshalYaml converts the input to yaml.
func MarshalYaml(in interface{}) ([]byte, error) {
	return Marshal(Yaml, in)
}

// MarshalJson converts the input struct to json.
func MarshalJson(in interface{}) ([]byte, error) {
	return Marshal(Json, in)
}

// MarshalToml converts the input to toml.
func MarshalToml(in interface{}) ([]byte, error) {
	return Marshal(Toml, in)
}

// Marshal a struct to the specified format.
// Supported formats are: yaml, json, and toml.
func Marshal(format string, in interface{}) (data []byte, err error) {
	switch format {
	case "json":
		data, err = json.MarshalIndent(in, "", "  ")
	case "yaml", "yml":
		w := bytes.Buffer{}
		encoder := yaml.NewEncoder(&w)
		encoder.SetIndent(2)
		err = encoder.Encode(in)
		data = w.Bytes()
	case "toml":
		data, err = toml.Marshal(in)
	default:
		return nil, newUnsupportedFormatError(format)
	}

	return data, errors.Wrapf(err, "error marshaling to %s", format)
}

// Unmarshal from the specified file into a struct.
// Supported file extensions are: yaml, yml, json, and toml.
func UnmarshalFile(fs aferox.Aferox, path string, out interface{}) error {
	data, err := fs.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "error reading file %s", path)
	}
	format := strings.TrimPrefix(filepath.Ext(path), ".")
	return Unmarshal(format, data, out)
}

// UnmarshalYaml converts the input yaml to a struct.
func UnmarshalYaml(data []byte, out interface{}) error {
	return Unmarshal(Yaml, data, out)
}

// UnmarshalJson converts the input json to a struct.
func UnmarshalJson(data []byte, out interface{}) error {
	return Unmarshal(Json, data, out)
}

// UnmarshalToml converts the input toml to a struct.
func UnmarshalToml(data []byte, out interface{}) error {
	return Unmarshal(Toml, data, out)
}

// Unmarshal from the specified format into a struct.
// Supported formats are: yaml, json, and toml.
func Unmarshal(format string, data []byte, out interface{}) error {
	switch format {
	case "json":
		return json.Unmarshal(data, out)
	case "yaml", "yml":
		return yaml.Unmarshal(data, out)
	case "toml":
		return toml.Unmarshal(data, out)
	default:
		return newUnsupportedFormatError(format)
	}
}

func newUnsupportedFormatError(format string) error {
	return errors.Errorf("unsupported format %s. Supported formats are: yaml, json and toml.", format)
}
