package porter

import (
	"fmt"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPorter_PrintMixins(t *testing.T) {
	p := NewTestPorter(t)

	opts := PrintMixinsOptions{
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatTable,
		},
	}
	err := p.PrintMixins(opts)

	require.Nil(t, err)
	wantOutput := `Name   Version   Author
exec   v1.0      Porter Authors
`
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Equal(t, wantOutput, gotOutput)
}

func TestPorter_InstallMixin(t *testing.T) {
	p := NewTestPorter(t)

	opts := mixin.InstallOptions{}
	opts.Name = "exec"
	opts.URL = "https://example.com"

	err := p.InstallMixin(opts)

	require.NoError(t, err)

	wantOutput := "installed exec mixin v1.0 (abc123)\n"
	gotOutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotOutput)
}

func TestPorter_UninstallMixin(t *testing.T) {
	p := NewTestPorter(t)

	opts := pkgmgmt.UninstallOptions{}
	err := opts.Validate([]string{"exec"})
	require.NoError(t, err, "Validate failed")

	err = p.UninstallMixin(opts)
	require.NoError(t, err, "UninstallMixin failed")

	wantOutput := "Uninstalled exec mixin"
	gotoutput := p.TestConfig.TestContext.GetOutput()
	assert.Contains(t, wantOutput, gotoutput)
}

func TestPorter_SearchMixins(t *testing.T) {
	fullList, err := ioutil.ReadFile("../mixin/directory/index.json")
	require.NoError(t, err, "could not read mixin directory")

	testcases := []struct {
		name       string
		mixin      string
		format     printer.Format
		wantOutput string
		wantErr    string
	}{{
		name:       "no name provided",
		mixin:      "",
		format:     printer.FormatJson,
		wantOutput: fmt.Sprintf("%s\n", string(fullList)),
	}, {
		name:   "mixin name single match",
		mixin:  "az",
		format: printer.FormatYaml,
		wantOutput: `- name: az
  author: Porter Authors
  description: A mixin for using the az cli
  url: https://cdn.porter.sh/mixins/atom.xml

`,
	}, {
		name:   "mixin name multiple match",
		mixin:  "ku",
		format: printer.FormatTable,
		wantOutput: `Name         Description                           Author           URL                                                                 URL Type
kubernetes   A mixin for using the kubectl cli     Porter Authors   https://cdn.porter.sh/mixins/atom.xml                               Atom Feed
kustomize    A mixin for using the kustomize cli   Don Stewart      https://github.com/donmstewart/porter-kustomize/releases/download   Download
`,
	}, {
		name:    "mixin name no match",
		mixin:   "ottersay",
		format:  printer.FormatYaml,
		wantErr: "no mixins found for ottersay",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			p.TestConfig.SetupPorterHome()

			opts := SearchOptions{
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
				Name: tc.mixin,
				Type: "mixin",
				List: mixin.GetMixinDirectory(),
			}

			err := p.SearchPackages(opts)
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
