package porter

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/config"
	"github.com/stretchr/testify/require"
)

func Test_generateInternalManifest(t *testing.T) {
	testcases := []struct {
		name         string
		opts         BuildOptions
		wantManifest string
	}{{
		name:         "no opts",
		opts:         BuildOptions{},
		wantManifest: "original.yaml",
	}, {
		name: "--file set",
		opts: BuildOptions{
			bundleFileOptions: bundleFileOptions{
				File: "alternate.yaml",
			},
		},
		wantManifest: "original.yaml",
	}, {
		name:         "name set",
		opts:         BuildOptions{metadataOpts: metadataOpts{Name: "newname"}},
		wantManifest: "new-name.yaml",
	}, {
		name:         "version set",
		opts:         BuildOptions{metadataOpts: metadataOpts{Version: "1.0.0"}},
		wantManifest: "new-version.yaml",
	}, {
		name:         "name and value set",
		opts:         BuildOptions{metadataOpts: metadataOpts{Name: "newname", Version: "1.0.0"}},
		wantManifest: "all-fields.yaml",
	}}

	p := NewTestPorter(t)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := config.Name
			if tc.opts.File != "" {
				manifest = tc.opts.File
			}
			p.TestConfig.TestContext.AddTestFile("testdata/generateManifest/original.yaml", manifest)

			err := tc.opts.Validate(p.Porter)
			require.NoError(t, err)

			err = p.generateInternalManifest(tc.opts)
			require.NoError(t, err)

			want := p.TestConfig.TestContext.AddTestFile(
				filepath.Join("testdata/generateManifest", tc.wantManifest), tc.wantManifest)

			got, err := p.FileSystem.ReadFile(build.LOCAL_MANIFEST)
			require.NoError(t, err)
			require.Equal(t, want, got)
		})
	}
}
