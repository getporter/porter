package pkgmgmt

import (
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/require"
)

func TestSearch_TestBox(t *testing.T) {
	fullList := PackageList{
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
		pkg       string
		wantItems PackageList
		wantError string
	}{{
		name:      "no args",
		pkg:       "",
		wantItems: fullList,
	}, {
		name:      "package name single match",
		pkg:       "az",
		wantItems: PackageList{fullList[0]},
	}, {
		name:      "package name multiple match",
		pkg:       "cowsay",
		wantItems: PackageList{fullList[1], fullList[2]},
	}, {
		name:      "package name no match",
		pkg:       "ottersay",
		wantItems: PackageList{},
		wantError: "no mixins found for ottersay",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			box := packr.Folder("./testdata/directory")
			searcher := NewSearcher(box)

			result, err := searcher.Search(tc.pkg, "mixin")
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.wantItems, result)
		})
	}
}
