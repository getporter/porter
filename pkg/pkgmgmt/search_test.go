package pkgmgmt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
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
			URL:         "https://github.com/carolynvs/porter-cowsay/releases/download",
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
		name:      "package name case insensitive",
		pkg:       "AZ",
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
			data, err := ioutil.ReadFile("testdata/directory/index.json")
			require.NoError(t, err)

			var pkgList PackageList
			err = json.Unmarshal(data, &pkgList)
			require.NoError(t, err)

			searcher := NewSearcher(pkgList)

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

func TestGetPackageListings_404(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	_, err := GetPackageListings(ts.URL)
	require.EqualError(t, err,
		fmt.Sprintf("unable to fetch package list via %s: Not Found", ts.URL))
}

func TestGetPackageListings_UnmarshalErr(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "foo")
	}))
	defer ts.Close()

	_, err := GetPackageListings(ts.URL)
	require.EqualError(t, err,
		"unable to unmarshal package list: invalid character 'o' in literal false (expecting 'a')")
}

func TestGetPackageListings_Success(t *testing.T) {
	packageList := PackageList{
		{
			Name:        "quokkasay",
			Author:      "Setonix Inc.",
			Description: "A mixin for using the quokkasay CLI",
			URL:         "https://cdn.quokkas.au/mixins/atom.xml",
		},
	}

	bytes, err := json.Marshal(packageList)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(bytes))
	}))
	defer ts.Close()

	list, err := GetPackageListings(ts.URL)
	require.NoError(t, err)
	require.Equal(t, packageList, list)
}
