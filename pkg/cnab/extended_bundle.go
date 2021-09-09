package cnab

import (
	"fmt"

	"get.porter.sh/porter/pkg/context"
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

// LoadBundle from the specified filepath.
func LoadBundle(c *context.Context, bundleFile string) (ExtendedBundle, error) {
	bunD, err := c.FileSystem.ReadFile(bundleFile)
	if err != nil {
		return ExtendedBundle{}, errors.Wrapf(err, "cannot read bundle at %s", bundleFile)
	}

	bun, err := bundle.Unmarshal(bunD)
	if err != nil {
		return ExtendedBundle{}, errors.Wrapf(err, "cannot load bundle from\n%s at %s", string(bunD), bundleFile)
	}

	return ExtendedBundle{*bun}, nil
}

// IsPorterBundle determines if the bundle was created by Porter.
func (b ExtendedBundle) IsPorterBundle() bool {
	_, madeByPorter := b.Custom[PorterExtension]
	return madeByPorter
}

// IsInternalParameter determines if the provided param is an internal parameter
// to Porter after analyzing the provided bundle
func (b ExtendedBundle) IsInternalParameter(param string) bool {
	if param, exists := b.Parameters[param]; exists {
		if def, exists := b.Definitions[param.Definition]; exists {
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
