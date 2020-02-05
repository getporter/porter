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

func TestSearch_AllResults(t *testing.T) {
	opts := SearchOptions{}

	box := packr.Folder("./remote-mixins")
	searcher := NewSearcher(box)

	// Not sure if this is valuable to track the number of results
	// Wanted to be sure we are somehow testing the 'prod' listing
	result, err := searcher.Search(opts)
	require.NoError(t, err)
	require.Equal(t, 12, len(result))
}

func TestSearch_TestBox(t *testing.T) {
	fullListing := []RemoteMixinInfo{
		{
			Name:        "az",
			Author:      "Porter Authors",
			Description: "A mixin for using the az cli",
			SourceURL:   "https://cdn.porter.sh/mixins/az",
			FeedURL:     "https://cdn.porter.sh/mixins/atom.xml",
		},
		{
			Name:        "cowsay",
			Author:      "Porter Authors",
			Description: "A mixin for using the cowsay cli",
			SourceURL:   "https://github.com/deislabs/porter-cowsay/releases/download",
			FeedURL:     "",
		},
		{
			Name:        "cowsayeth",
			Author:      "Udder Geniuses",
			Description: "A mixin for using the cowsayeth cli",
			SourceURL:   "https://cdn.uddergenius.es/mixins/cowsayeth",
			FeedURL:     "https://cdn.uddergenius.es/mixins/atom.xml",
		},
	}

	testcases := []struct {
		name      string
		opts      SearchOptions
		wantItems []RemoteMixinInfo
	}{
		{"no args",
			SearchOptions{},
			fullListing,
		},
		{"mixin name single match",
			SearchOptions{Name: "az"},
			[]RemoteMixinInfo{fullListing[0]},
		},
		{"mixin name multiple match",
			SearchOptions{Name: "cowsay"},
			[]RemoteMixinInfo{fullListing[1], fullListing[2]},
		},
		{"mixin name no match",
			SearchOptions{Name: "ottersay"},
			[]RemoteMixinInfo{},
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
