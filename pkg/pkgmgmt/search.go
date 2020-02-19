package pkgmgmt

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
)

// Searcher contains a packr.Box containing a searchable list of packages
type Searcher struct {
	Box *packr.Box
}

// NewSearcher returns a Searcher with the provided packr.Box
func NewSearcher(box *packr.Box) Searcher {
	return Searcher{
		Box: box,
	}
}

// Search searches for packages matching the optional provided name,
// returning the full list if none is provided
func (s *Searcher) Search(name, pkgType string) (PackageList, error) {
	data, err := s.Box.Find("index.json")
	if err != nil {
		return PackageList{}, errors.Wrapf(err, "error loading %s list\n", pkgType)
	}

	var pl PackageList
	err = json.Unmarshal(data, &pl)
	if err != nil {
		return PackageList{}, errors.Wrapf(err, "could not parse %s list\n", pkgType)
	}

	if name == "" {
		sort.Sort(pl)
		return pl, nil
	}

	results := PackageList{}
	query := strings.ToLower(name)
	for _, p := range pl {
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
