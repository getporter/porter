package porter

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchOptions_Validate(t *testing.T) {
	opts := SearchOptions{
		Type: "plugin",
		PrintOptions: printer.PrintOptions{
			RawFormat: "json",
		},
	}
	args := []string{}

	err := opts.Validate(args)
	require.NoError(t, err)

	opts.Type = "mixin"
	err = opts.Validate(args)
	require.NoError(t, err)

	opts.Type = "mixxin"
	err = opts.Validate(args)
	require.EqualError(t, err, "unsupported package type: mixxin")
}

func TestSearchOptions_Validate_PackageName(t *testing.T) {
	opts := SearchOptions{}

	err := opts.validatePackageName([]string{})
	require.NoError(t, err)
	assert.Equal(t, "", opts.Name)

	err = opts.validatePackageName([]string{"helm"})
	require.NoError(t, err)
	assert.Equal(t, "helm", opts.Name)

	err = opts.validatePackageName([]string{"helm", "nstuff"})
	require.EqualError(t, err, "only one positional argument may be specified, the package name, but multiple were received: [helm nstuff]")
}

func TestPorter_SearchPackages_Mixins(t *testing.T) {
	testcases := []struct {
		name               string
		mixin              string
		format             printer.Format
		wantOutput         string
		wantNonEmptyOutput bool
		wantErr            string
	}{{
		name:               "no name provided",
		mixin:              "",
		format:             printer.FormatJson,
		wantNonEmptyOutput: true,
	}, {
		name:       "mixin name single match",
		mixin:      "az",
		format:     printer.FormatYaml,
		wantOutput: "testdata/packages/search-single-match.txt",
	}, {
		name:       "mixin name multiple match",
		mixin:      "ku",
		format:     printer.FormatPlaintext,
		wantOutput: "testdata/packages/search-multi-match.txt",
	}, {
		name:    "mixin name no match",
		mixin:   "ottersay",
		format:  printer.FormatYaml,
		wantErr: "no mixins found for ottersay",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			opts := SearchOptions{
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
				Name: tc.mixin,
				Type: "mixin",
			}

			err := p.SearchPackages(opts)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				gotOutput := p.TestConfig.TestContext.GetOutput()

				// Only check that the output isn't empty, but don't try to match the exact contents because it changes
				if tc.wantNonEmptyOutput {
					assert.NotEmpty(t, gotOutput, "expected the output to not be empty")
				} else {
					test.CompareGoldenFile(t, tc.wantOutput, gotOutput)
				}
			}
		})
	}
}

func TestPorter_SearchPackages_Plugins(t *testing.T) {
	// Fetch the full plugin list for comparison in test case(s)
	fullList, err := fetchFullListBytes("plugin")
	require.NoError(t, err)

	testcases := []struct {
		name       string
		plugin     string
		format     printer.Format
		wantOutput string
		wantErr    string
	}{{
		name:       "no name provided",
		plugin:     "",
		format:     printer.FormatJson,
		wantOutput: fmt.Sprintf("%s\n", string(fullList)),
	}, {
		name:   "plugin name single match",
		plugin: "az",
		format: printer.FormatYaml,
		wantOutput: `- name: azure
  author: Porter Authors
  description: Integrate Porter with Azure. Store Porter's data in Azure Cloud and secure your bundle's secrets in Azure Key Vault.
  url: https://cdn.porter.sh/plugins/atom.xml
`,
	}, {
		name:    "plugin name no match",
		plugin:  "ottersay",
		format:  printer.FormatYaml,
		wantErr: "no plugins found for ottersay",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			opts := SearchOptions{
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
				Name: tc.plugin,
				Type: "plugin",
			}

			err = p.SearchPackages(opts)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			gotOutput := p.TestConfig.TestContext.GetOutput()
			require.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}

// fetchFullListBytes fetches the full package list according to the
// provided package type, sorts the list, and returns its marshaled byte form
func fetchFullListBytes(pkgType string) ([]byte, error) {
	url := pkgmgmt.GetPackageListURL(pkgmgmt.GetDefaultPackageMirrorURL(), pkgType)
	packageList, err := pkgmgmt.GetPackageListings(url)
	if err != nil {
		return nil, err
	}

	sort.Sort(packageList)
	bytes, err := json.MarshalIndent(packageList, "", "  ")
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
