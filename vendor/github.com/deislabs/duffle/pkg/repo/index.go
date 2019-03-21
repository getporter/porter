package repo

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sort"

	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrNoAPIVersion indicates that an API version was not specified.
	ErrNoAPIVersion = errors.New("no API version specified")
	// ErrNoBundleVersion indicates that a bundle with the given version is not found.
	ErrNoBundleVersion = errors.New("no bundle with the given version found")
	// ErrNoBundleName indicates that a bundle with the given name is not found.
	ErrNoBundleName = errors.New("no bundle name found")
)

type BundleVersion struct {
	Version *semver.Version
	Digest  string
}

// ByVersion implements sort.Interface for []BundleVersion based on
// the version field.
type ByVersion []BundleVersion

func (a ByVersion) Len() int           { return len(a) }
func (a ByVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVersion) Less(i, j int) bool { return a[i].Version.LessThan(a[j].Version) }

// Index defines a list of bundle repositories, each repository's respective tags and the digest reference.
type Index map[string]map[string]string

// LoadIndex takes a file at the given path and returns an Index object
func LoadIndex(path string) (Index, error) {
	f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadIndex(f)
}

// LoadIndexReader takes a reader and returns an Index object
func LoadIndexReader(r io.Reader) (Index, error) {
	return loadIndex(r)
}

// LoadIndexBuffer reads repository metadata from a JSON byte stream
func LoadIndexBuffer(data []byte) (Index, error) {
	return loadIndex(bytes.NewBuffer(data))
}

// Add adds a new entry to the index
func (i Index) Add(name, version string, digest string) {
	if tags, ok := i[name]; ok {
		tags[version] = digest
	} else {
		i[name] = map[string]string{
			version: digest,
		}
	}
}

// Delete removes a bundle from the index.
//
// Returns false if no record was found to delete.
func (i Index) Delete(name string) bool {
	_, ok := i[name]
	if ok {
		delete(i, name)
	}
	return ok
}

// DeleteVersion removes a single version of a given bundle from the index.
//
// Returns false if the name or version is not found.
func (i Index) DeleteVersion(name, version string) bool {
	sub, ok := i[name]
	if !ok {
		return false
	}
	_, ok = sub[version]
	if ok {
		delete(sub, version)
	}
	return ok
}

// Has returns true if the index has an entry for a bundle with the given name and exact version.
func (i Index) Has(name, version string) bool {
	_, err := i.Get(name, version)
	return err == nil
}

// Get returns the digest for the given name.
//
// If version is empty, this will return the digest for the bundle with the highest version.
func (i Index) Get(name, version string) (string, error) {
	var versions ByVersion
	versions, ok := i.GetVersions(name)
	if !ok {
		return "", ErrNoBundleName
	}
	if len(versions) == 0 {
		return "", ErrNoBundleVersion
	}

	sort.Sort(sort.Reverse(versions))

	var constraint *semver.Constraints
	if len(version) == 0 {
		constraint, _ = semver.NewConstraint("*")
	} else {
		var err error
		constraint, err = semver.NewConstraint(version)
		if err != nil {
			return "", err
		}
	}

	for _, ver := range versions {
		if constraint.Check(ver.Version) {
			return ver.Digest, nil
		}
	}
	return "", ErrNoBundleVersion
}

// GetVersions gets all of the versions for the given name.
//
// If the name is not found, this will return false.
func (i Index) GetVersions(name string) ([]BundleVersion, bool) {
	ret, ok := i[name]
	if !ok {
		ret, ok = i.versionsWithDigest(name)
	}

	var bv []BundleVersion
	for version, digest := range ret {
		v, err := semver.NewVersion(version)
		if err != nil {
			log.Debugf("found a version in the index that is not semver compatible: '%s'\n", version)
			continue
		}

		bv = append(bv, BundleVersion{Version: v, Digest: digest})
	}

	return bv, ok
}

func (i Index) versionsWithDigest(digest string) (map[string]string, bool) {
	for _, versions := range i {
		for v, d := range versions {
			if d == digest {
				return map[string]string{v: d}, true
			}
		}
	}

	return nil, false
}

// WriteFile writes an index file to the given destination path.
//
// The mode on the file is set to 'mode'.
func (i Index) WriteFile(dest string, mode os.FileMode) error {
	b, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dest, b, mode)
}

// Merge merges the src index into i (dest).
//
// This merges by name and version.
//
// If one of the entries in the destination index does _not_ already exist, it is added.
// In all other cases, the existing record is preserved.
func (i *Index) Merge(src Index) {
	for name, versionMap := range src {
		for version, digest := range versionMap {
			if !i.Has(name, version) {
				i.Add(name, version, digest)
			}
		}
	}
}

// loadIndex loads an index file and does minimal validity checking.
func loadIndex(r io.Reader) (Index, error) {
	i := Index{}
	if err := json.NewDecoder(r).Decode(&i); err != nil && err != io.EOF {
		return i, err
	}
	return i, nil
}
