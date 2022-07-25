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


func MakeCNABCompatible(schema *definition.Schema) {
	if v, ok := schema.Type.(string); ok {
		if t, ok := config.PorterParamMap[v]; ok {
			schema.Type = t;
			schema.ContentEncoding = "base64"
		}
	}
}