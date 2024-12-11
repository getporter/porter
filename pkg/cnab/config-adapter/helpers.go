package configadapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
)

func LoadTestBundle(t *testing.T, config *config.Config, path string) cnab.ExtendedBundle {
	ctx := context.Background()
	m, err := manifest.ReadManifest(config.Context, path, config)
	require.NoError(t, err)
	b, err := ConvertToTestBundle(ctx, config, m)
	require.NoError(t, err)
	return b
}

// ConvertToTestBundle is suitable for taking a test manifest (porter.yaml)
// and making a bundle.json for it. Does not make an accurate representation
// of the bundle, but is suitable for testing.
func ConvertToTestBundle(ctx context.Context, cfg *config.Config, manifest *manifest.Manifest) (cnab.ExtendedBundle, error) {
	converter := NewManifestConverter(cfg, manifest, nil, nil, false)
	return converter.ToBundle(ctx)
}
