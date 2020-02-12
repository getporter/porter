package mixin

import (
	"testing"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchOptions_Validate_MixinName(t *testing.T) {
	opts := SearchOptions{}

	err := opts.validateMixinName([]string{})
	require.NoError(t, err)
	assert.Equal(t, "", opts.Name)

	err = opts.validateMixinName([]string{"helm"})
	require.NoError(t, err)
	assert.Equal(t, "helm", opts.Name)

	err = opts.validateMixinName([]string{"helm", "nstuff"})
	require.EqualError(t, err, "only one positional argument may be specified, the mixin name, but multiple were received: [helm nstuff]")
}

func TestSearch_TestBox(t *testing.T) {
	fullList := pkgmgmt.PackageList{
		{
			Name:        "az",
			Author:      "Porter Authors",
			Description: "A mixin for using the az cli",
			URL:         "https://cdn.porter.sh/mixins/atom.xml",
		},
		{
			Name:        "cowsay",
			Author:      "Porter Authors",
			Description: "A mixin for using the cowsay cli",
			URL:         "https://github.com/deislabs/porter-cowsay/releases/download",
		},
		{
			Name:        "cowsayeth",
			Author:      "Udder Geniuses",
			Description: "A mixin for using the cowsayeth cli",
			URL:         "https://cdn.uddergenius.es/mixins/atom.xml",
		},
	}

	testcases := []struct {
		name      string
		opts      SearchOptions
		wantItems pkgmgmt.PackageList
		wantError string
	}{{
		name:      "no args",
		opts:      SearchOptions{},
		wantItems: fullList,
	}, {
		name:      "mixin name single match",
		opts:      SearchOptions{Name: "az"},
		wantItems: pkgmgmt.PackageList{fullList[0]},
	}, {
		name:      "mixin name multiple match",
		opts:      SearchOptions{Name: "cowsay"},
		wantItems: pkgmgmt.PackageList{fullList[1], fullList[2]},
	}, {
		name:      "mixin name no match",
		opts:      SearchOptions{Name: "ottersay"},
		wantItems: pkgmgmt.PackageList{},
		wantError: "no mixins found for ottersay",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			box := packr.Folder("./testdata/directory")
			searcher := NewSearcher(box)

			result, err := searcher.Search(tc.opts)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.wantItems, result)
		})
	}
}
