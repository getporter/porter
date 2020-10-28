package pkgmgmt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// Searcher contains a packr.Box containing a searchable list of packages
type Searcher struct {
	List PackageList
}

// NewSearcher returns a Searcher with the provided packr.Box
func NewSearcher(list PackageList) Searcher {
	return Searcher{
		List: list,
	}
}

// Search searches for packages matching the optional provided name,
// returning the full list if none is provided
func (s *Searcher) Search(name, pkgType string) (PackageList, error) {
	if name == "" {
		sort.Sort(s.List)
		return s.List, nil
	}

	results := PackageList{}
	query := strings.ToLower(name)
	for _, p := range s.List {
		if strings.Contains(p.Name, query) {
			results = append(results, p)
		}
	}

	if results.Len() == 0 {
		return PackageList{}, errors.Errorf("no %ss found for %s", pkgType, name)
	}

	sort.Sort(results)
	return results, nil
}

// GetPackageListings returns the listings for packages via the provided URL
func GetPackageListings(url string) (PackageList, error) {
	resp, err := http.Get(url)
	if err != nil {
		return PackageList{}, errors.Wrapf(err, "unable to fetch package list via %s", url)
	}
	if resp.StatusCode != http.StatusOK {
		return PackageList{}, fmt.Errorf("unable to fetch package list via %s: %s", url, http.StatusText(resp.StatusCode))
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return PackageList{}, errors.Wrapf(err, "unable to read package list via %s", url)
	}

	list := PackageList{}
	err = json.Unmarshal(data, &list)
	if err != nil {
		return PackageList{}, errors.Wrap(err, "unable to unmarshal package list")
	}

	return list, nil
}
