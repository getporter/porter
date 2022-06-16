package cnab

import (
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/pkg/errors"
)

// ExtendedBundle is a bundle that has typed access to extensions declared in the bundle,
// allowing quick type-safe access to custom extensions from the CNAB spec.
type ExtendedBundle struct {
	bundle.Bundle
}

// NewBundle creates an ExtendedBundle from a given bundle.
func NewBundle(bundle bundle.Bundle) ExtendedBundle {
	return ExtendedBundle{bundle}
}

// LoadBundle from the specified filepath.
func LoadBundle(c *portercontext.Context, bundleFile string) (ExtendedBundle, error) {
	bunD, err := c.FileSystem.ReadFile(bundleFile)
	if err != nil {
		return ExtendedBundle{}, errors.Wrapf(err, "cannot read bundle at %s", bundleFile)
	}

	bun, err := bundle.Unmarshal(bunD)
	if err != nil {
		return ExtendedBundle{}, errors.Wrapf(err, "cannot load bundle from\n%s at %s", string(bunD), bundleFile)
	}

	return NewBundle(*bun), nil
}

// IsPorterBundle determines if the bundle was created by Porter.
func (b ExtendedBundle) IsPorterBundle() bool {
	_, madeByPorter := b.Custom[PorterExtension]
	return madeByPorter
}

// IsInternalParameter determines if the provided parameter is internal
// to Porter after analyzing the provided bundle.
func (b ExtendedBundle) IsInternalParameter(name string) bool {
	if param, exists := b.Parameters[name]; exists {
		if def, exists := b.Definitions[param.Definition]; exists {
			return def.Comment == PorterInternal
		}
	}
	return false
}

// IsInternalOutput determines if the provided output is internal
// to Porter after analyzing the provided bundle.
func (b ExtendedBundle) IsInternalOutput(name string) bool {
	if output, exists := b.Outputs[name]; exists {
		if def, exists := b.Definitions[output.Definition]; exists {
			return def.Comment == PorterInternal
		}
	}
	return false
}

// IsSensitiveParameter determines if the parameter contains a sensitive value.
func (b ExtendedBundle) IsSensitiveParameter(param string) bool {
	if param, exists := b.Parameters[param]; exists {
		if def, exists := b.Definitions[param.Definition]; exists {
			return def.WriteOnly != nil && *def.WriteOnly
		}
	}
	return false
}

// GetParameterType determines the type of parameter accounting for
// Porter-specific parameter types like file.
func (b ExtendedBundle) GetParameterType(def *definition.Schema) string {
	if b.IsFileType(def) {
		return "file"
	}

	if def.ID == claim.OutputInvocationImageLogs {
		return "string"
	}

	return fmt.Sprintf("%v", def.Type)
}

// IsFileType determines if the parameter/credential is of type "file".
func (b ExtendedBundle) IsFileType(def *definition.Schema) bool {
	return b.SupportsFileParameters() &&
		def.Type == "string" && def.ContentEncoding == "base64"
}

// ConvertParameterValue converts a parameter's value from an unknown type,
// it could be a string from stdin or another Go type, into the type of the
// parameter as defined in the bundle.
func (b ExtendedBundle) ConvertParameterValue(key string, value interface{}) (interface{}, error) {
	param, ok := b.Parameters[key]
	if !ok {
		return nil, fmt.Errorf("unable to convert the parameters' value to the destination parameter type because parameter %s not defined in bundle", key)
	}

	def, ok := b.Definitions[param.Definition]
	if !ok {
		return nil, fmt.Errorf("unable to convert the parameters' value to the destination parameter type because parameter %s has no definition", key)
	}

	if def.Type != nil {
		switch t := value.(type) {
		case string:
			typedValue, err := def.ConvertValue(t)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to convert parameter's %s value %s to the destination parameter type %s", key, value, def.Type)
			}
			return typedValue, nil
		case json.Number:
			switch def.Type {
			case "integer":
				return t.Int64()
			case "number":
				return t.Float64()
			default:
				return t.String(), nil
			}
		default:
			return t, nil
		}
	} else {
		return value, nil
	}
}

func (b ExtendedBundle) WriteParameterToString(paramName string, value interface{}) (string, error) {
	return WriteParameterToString(paramName, value)
}

// WriteParameterToString changes a parameter's value from its type as
// defined by the bundle to its runtime string representation.
// The value should have already been converted to its bundle representation
// by calling ConvertParameterValue.
func WriteParameterToString(paramName string, value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}

	if stringVal, ok := value.(string); ok {
		return stringVal, nil
	}

	contents, err := json.Marshal(value)
	return string(contents), errors.Wrapf(err, "could not marshal the value for parameter %s to a json string: %#v", paramName, value)
}
