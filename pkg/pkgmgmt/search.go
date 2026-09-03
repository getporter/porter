package pkgmgmt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Searcher can locate a mixin or plugin from the community feeds.
type Searcher struct {
	List PackageList
}

// NewSearcher creates a new Searcher from a package distribution list.
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
		return PackageList{}, fmt.Errorf("no %ss found for %s", pkgType, name)
	}

	sort.Sort(results)
	return results, nil
}

// GetPackageListings returns the listings for packages via the provided URL
func GetPackageListings(url string) (PackageList, error) {
	const maxRetries = 3
	const baseDelay = 1 * time.Second

	var lastErr error
	var resp *http.Response
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(baseDelay * time.Duration(1<<uint(attempt-1)))
		}

		var err error
		resp, err = http.Get(url)
		if err != nil {
			lastErr = err
			resp = nil
			if IsRetryableError(err) {
				continue
			}
			return PackageList{}, fmt.Errorf("unable to fetch package list via %s: %w", url, err)
		}

		if resp.StatusCode != http.StatusOK {
			statusCode := resp.StatusCode
			resp.Body.Close()
			resp = nil
			lastErr = fmt.Errorf("unable to fetch package list via %s: %s", url, http.StatusText(statusCode))
			if IsRetryableStatus(statusCode) {
				continue
			}
			return PackageList{}, lastErr
		}

		break
	}

	if resp == nil {
		return PackageList{}, fmt.Errorf("failed to fetch package list via %s after %d attempts: %w", url, maxRetries, lastErr)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return PackageList{}, fmt.Errorf("unable to read package list via %s: %w", url, err)
	}

	list := PackageList{}
	err = json.Unmarshal(data, &list)
	if err != nil {
		return PackageList{}, fmt.Errorf("unable to unmarshal package list: %w", err)
	}

	return list, nil
}
