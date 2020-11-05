package porter

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestUpdateManifest(t *testing.T) {
	testcases := []struct {
		name         string
		opts         updateManifestOpts
		wantManifest string
	}{{
		name:         "no opts",
		opts:         updateManifestOpts{},
		wantManifest: "original.yaml",
	}, {
		name:         "name set",
		opts:         updateManifestOpts{Name: "new name"},
		wantManifest: "new-name.yaml",
	}, {
		name:         "version set",
		opts:         updateManifestOpts{Version: "1.0.0"},
		wantManifest: "new-version.yaml",
	}, {
		name:         "name and value set",
		opts:         updateManifestOpts{Name: "new name", Version: "1.0.0"},
		wantManifest: "all-fields.yaml",
	}}

	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.TestConfig.TestContext.AddTestFile("testdata/update-manifest/original.yaml", config.Name)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := p.updateManifest(config.Name, tc.opts)
			require.NoError(t, err)

			want := p.TestConfig.TestContext.AddTestFile(
				filepath.Join("testdata/update-manifest", tc.wantManifest), tc.wantManifest)

			got, err := p.FileSystem.ReadFile(build.LOCAL_MANIFEST)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}
