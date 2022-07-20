package configadapter

import (
	"context"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/cnabio/cnab-go/bundle/definition"
)

// ConvertToTestBundle is suitable for taking a test manifest (porter.yaml)
// and making a bundle.json for it. Does not make an accurate representation
// of the bundle, but is suitable for testing.
func ConvertToTestBundle(ctx context.Context, cfg *config.Config, manifest *manifest.Manifest) (cnab.ExtendedBundle, error) {
	converter := NewManifestConverter(cfg, manifest, nil, nil)
	return converter.ToBundle(ctx)
}

// MakeCNABCompatible receives a schema with possible porter specific parameters
// and converts those parameters to CNAB compatible versions.
// Returns true if values were replaced and false otherwise.
func MakeCNABCompatible(schema *definition.Schema) bool {
	if v, ok := schema.Type.(string); ok {
		if c, ok := config.PorterParamMap[v]; ok {
			schema.Type = c.Type
			schema.ContentEncoding = c.Encoding
			schema.Comment = c.Comment
			return ok
		}
	}

	return false
}