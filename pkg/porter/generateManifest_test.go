package porter

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestgenerateInternalManifest(t *testing.T) {
	testcases := []struct {
		name         string
		opts         metadataOpts
		wantManifest string
	}{{
		name:         "no opts",
		opts:         metadataOpts{},
		wantManifest: "original.yaml",
	}, {
		name:         "name set",
		opts:         metadataOpts{Name: "newname"},
		wantManifest: "new-name.yaml",
	}, {
		name:         "version set",
		opts:         metadataOpts{Version: "1.0.0"},
		wantManifest: "new-version.yaml",
	}, {
		name:         "name and value set",
		opts:         metadataOpts{Name: "newname", Version: "1.0.0"},
		wantManifest: "all-fields.yaml",
	}}

	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.TestConfig.TestContext.AddTestFile("testdata/generateManifest/original.yaml", config.Name)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := p.generateInternalManifest(tc.opts)
			require.NoError(t, err)

			want := p.TestConfig.TestContext.AddTestFile(
				filepath.Join("testdata/generateManifest", tc.wantManifest), tc.wantManifest)

			got, err := p.FileSystem.ReadFile(build.LOCAL_MANIFEST)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}
