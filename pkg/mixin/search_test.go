package mixin

import (
	"testing"

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
	fullList := RemoteMixinList{
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
		wantItems RemoteMixinList
	}{
		{"no args",
			SearchOptions{},
			fullList,
		},
		{"mixin name single match",
			SearchOptions{Name: "az"},
			RemoteMixinList{fullList[0]},
		},
		{"mixin name multiple match",
			SearchOptions{Name: "cowsay"},
			RemoteMixinList{fullList[1], fullList[2]},
		},
		{"mixin name no match",
			SearchOptions{Name: "ottersay"},
			RemoteMixinList{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			box := packr.Folder("./testdata/remote-mixins")
			searcher := NewSearcher(box)

			result, err := searcher.Search(tc.opts)
			require.NoError(t, err)
			require.Equal(t, tc.wantItems, result)
		})
	}
}
