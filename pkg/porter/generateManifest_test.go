package porter

import (
	"context"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/test"
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
		wantManifest: "expected-result.yaml",
	}, {
		name: "--file set",
		opts: BuildOptions{
			bundleFileOptions: bundleFileOptions{
				File: "alternate.yaml",
			},
		},
		wantManifest: "expected-result.yaml",
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
	}, {
		name:         "custom input set",
		opts:         BuildOptions{Customs: []string{"key1=editedValue1", "key2.nestedKey2=editedValue2"}},
		wantManifest: "custom-input.yaml",
	}}

	p := NewTestPorter(t)
	defer p.Close()
	p.TestRegistry.MockPullImage = mockPullImage

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			manifest := config.Name
			if tc.opts.File != "" {
				manifest = tc.opts.File
			}
			p.TestConfig.TestContext.AddTestFile("testdata/generateManifest/original.yaml", manifest)

			err := tc.opts.Validate(p.Porter)
			require.NoError(t, err)

			err = p.generateInternalManifest(context.Background(), tc.opts)
			require.NoError(t, err)

			goldenFile := filepath.Join("testdata/generateManifest", tc.wantManifest)
			p.TestConfig.TestContext.AddTestFile(goldenFile, tc.wantManifest)
			got, err := p.FileSystem.ReadFile(build.LOCAL_MANIFEST)
			require.NoError(t, err)
			test.CompareGoldenFile(t, goldenFile, string(got))
		})
	}
}

func mockPullImage(ctx context.Context, image string) (string, error) {
	return "sha256:8b92b7269f59e3ed824e811a1ff1ee64f0d44c0218efefada57a4bebc2d7ef6f", nil
}
