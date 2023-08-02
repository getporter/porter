package porter

import (
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/build"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedOptions_defaultBundleFiles(t *testing.T) {
	cxt := portercontext.NewTestContext(t)

	_, err := cxt.FileSystem.Create("porter.yaml")
	require.NoError(t, err)

	opts := installationOptions{}
	err = opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, "porter.yaml", opts.File)
	assert.Equal(t, filepath.FromSlash(".cnab/bundle.json"), opts.CNABFile)
}

func TestSharedOptions_defaultBundleFiles_AltManifest(t *testing.T) {
	cxt := portercontext.NewTestContext(t)

	opts := installationOptions{
		BundleDefinitionOptions: BundleDefinitionOptions{
			File: "mybun/porter.yaml",
		},
	}
	err := opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, filepath.FromSlash(".cnab/bundle.json"), opts.CNABFile)
}

func TestSharedOptions_defaultBundleFiles_CNABFile(t *testing.T) {
	cxt := portercontext.NewTestContext(t)

	// Add existing porter manifest; ensure it isn't processed when cnab-file is spec'd
	_, err := cxt.FileSystem.Create("porter.yaml")
	require.NoError(t, err)
	_, err = cxt.FileSystem.Create("mycnabfile.json")
	require.NoError(t, err)

	opts := installationOptions{}
	opts.CNABFile = "mycnabfile.json"
	err = opts.defaultBundleFiles(cxt.Context)
	require.NoError(t, err)

	assert.Equal(t, "", opts.File)
	assert.Equal(t, "mycnabfile.json", opts.CNABFile)
}

func TestSharedOptions_validateBundleJson(t *testing.T) {
	cxt := portercontext.NewTestContext(t)

	cxt.FileSystem.Create("mybun1/bundle.json")
	cxt.FileSystem.Create("bundle1.json")

	testcases := []struct {
		name           string
		cnabFile       string
		wantBundleJson string
		wantError      string
	}{
		{name: "absolute file exists", cnabFile: "/mybun1/bundle.json", wantBundleJson: "/mybun1/bundle.json", wantError: ""},
		{name: "relative file exists", cnabFile: "bundle1.json", wantBundleJson: "/bundle1.json", wantError: ""},
		{name: "absolute file does not exist", cnabFile: "mybun2/bundle.json", wantError: "unable to access --cnab-file mybun2/bundle.json"},
		{name: "relative file does not", cnabFile: "bundle2.json", wantError: "unable to access --cnab-file bundle2.json"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := installationOptions{
				BundleDefinitionOptions: BundleDefinitionOptions{
					CNABFile: tc.cnabFile,
				},
			}

			err := opts.validateCNABFile(cxt.Context)

			if tc.wantError == "" {
				require.NoError(t, err)
				wantBundleJsonAbs, err := filepath.Abs(tc.wantBundleJson)
				require.NoError(t, err)

				assert.Equal(t, wantBundleJsonAbs, opts.CNABFile)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError)
			}
		})
	}
}

func Test_bundleFileOptions(t *testing.T) {
	testcases := []struct {
		name         string
		opts         BundleDefinitionOptions
		setup        func(*portercontext.Context, BundleDefinitionOptions) error
		wantFile     string
		wantCNABFile string
		wantError    string
	}{
		{
			name:         "no opts",
			opts:         BundleDefinitionOptions{},
			setup:        func(ctx *portercontext.Context, opts BundleDefinitionOptions) error { return nil },
			wantFile:     "/" + config.Name,
			wantCNABFile: "/" + build.LOCAL_BUNDLE,
			wantError:    "",
		}, {
			name: "reference set",
			opts: BundleDefinitionOptions{
				ReferenceSet: true,
			},
			setup:        func(ctx *portercontext.Context, opts BundleDefinitionOptions) error { return nil },
			wantFile:     "",
			wantCNABFile: "",
			wantError:    "",
		}, {
			name: "invalid dir",
			opts: BundleDefinitionOptions{
				Dir: "path/to/bundle",
			},
			setup:        func(ctx *portercontext.Context, opts BundleDefinitionOptions) error { return nil },
			wantFile:     "",
			wantCNABFile: "",
			wantError:    `"path/to/bundle" is not a valid directory: open /path/to/bundle: file does not exist`,
		}, {
			name: "invalid file",
			opts: BundleDefinitionOptions{
				File: "alternate/porter.yaml",
			},
			setup:        func(ctx *portercontext.Context, opts BundleDefinitionOptions) error { return nil },
			wantFile:     "",
			wantCNABFile: "",
			wantError:    "unable to access --file /alternate/porter.yaml: open /alternate/porter.yaml: file does not exist",
		}, {
			name: "valid dir",
			opts: BundleDefinitionOptions{
				Dir: "path/to/bundle",
			},
			setup: func(ctx *portercontext.Context, opts BundleDefinitionOptions) error {
				err := ctx.FileSystem.MkdirAll(filepath.Join(opts.Dir, config.Name), pkg.FileModeDirectory)
				if err != nil {
					return err
				}
				return ctx.FileSystem.MkdirAll(opts.Dir, pkg.FileModeDirectory)
			},
			wantFile:     "/path/to/bundle/porter.yaml",
			wantCNABFile: "/path/to/bundle/.cnab/bundle.json",
			wantError:    "",
		}, {
			name: "valid file",
			opts: BundleDefinitionOptions{
				File: "alternate/porter.yaml",
			},
			setup: func(ctx *portercontext.Context, opts BundleDefinitionOptions) error {
				return ctx.FileSystem.MkdirAll(opts.File, pkg.FileModeDirectory)
			},
			wantFile:     "/alternate/porter.yaml",
			wantCNABFile: "/" + build.LOCAL_BUNDLE,
			wantError:    "",
		}, {
			name: "valid dir and file",
			opts: BundleDefinitionOptions{
				Dir:  "path/to/bundle",
				File: "alternate/porter.yaml",
			},
			setup: func(ctx *portercontext.Context, opts BundleDefinitionOptions) error {
				err := ctx.FileSystem.MkdirAll(filepath.Join(opts.Dir, opts.File), pkg.FileModeDirectory)
				if err != nil {
					return err
				}
				return ctx.FileSystem.MkdirAll(opts.Dir, pkg.FileModeDirectory)
			},
			wantFile:     "/path/to/bundle/alternate/porter.yaml",
			wantCNABFile: "/path/to/bundle/.cnab/bundle.json",
			wantError:    "",
		}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cxt := portercontext.NewTestContext(t)

			// Create default local manifest
			_, err := cxt.FileSystem.Create(config.Name)
			require.NoError(t, err)

			err = tc.setup(cxt.Context, tc.opts)
			require.NoError(t, err)

			err = tc.opts.Validate(cxt.Context)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err)

				require.Equal(t, tc.wantFile, tc.opts.File)
				require.Equal(t, tc.wantCNABFile, tc.opts.CNABFile)

				// Working Dir assertions
				wd := cxt.FileSystem.Getwd()
				if tc.opts.Dir != "" && tc.wantError == "" {
					require.Equal(t, tc.opts.Dir, wd)
				} else {
					require.Equal(t, "/", wd)
				}
			}
		})
	}
}
