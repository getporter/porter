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
	fullList, err := ioutil.ReadFile("../mixin/remote-mixins/index.json")
	require.NoError(t, err, "could not read remote mixin list")

	testcases := []struct {
		name       string
		mixin      string
		format     printer.Format
		wantOutput string
	}{
		{"no name provided",
			"",
			printer.FormatJson,
			fmt.Sprintf("%s\n", string(fullList)),
		},
		{"mixin name single match",
			"az",
			printer.FormatYaml,
			`- name: az
  author: Porter Authors
  description: A mixin for using the az cli
  sourceurl: https://cdn.porter.sh/mixins/az
  feedurl: https://cdn.porter.sh/mixins/atom.xml

`,
		},
		{"mixin name multiple match",
			"ku",
			printer.FormatTable,
			`Name         Description                           Author           Source URL                                                          Feed URL
kubernetes   A mixin for using the kubectl cli     Porter Authors   https://cdn.porter.sh/mixins/kubernetes                             https://cdn.porter.sh/mixins/atom.xml
kustomize    A mixin for using the kustomize cli   Don Stewart      https://github.com/donmstewart/porter-kustomize/releases/download   
`,
		},
		{"mixin name no match",
			"ottersay",
			printer.FormatYaml,
			"No mixins found for ottersay\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			p.TestConfig.SetupPorterHome()

			opts := mixin.SearchOptions{
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
				Name: tc.mixin,
			}

			err := p.SearchMixins(opts)
			require.NoError(t, err)

			gotOutput := p.TestConfig.TestContext.GetOutput()
			require.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}
