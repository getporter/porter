package cnab

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/registry"
	"github.com/opencontainers/go-digest"
)

// ParseOCIReference parses the specified value as an OCIReference.
// If the reference includes a digest, the digest is validated.
func ParseOCIReference(value string) (OCIReference, error) {
	named, err := reference.ParseNormalizedNamed(value)
	if err != nil {
		return OCIReference{}, fmt.Errorf("failed to parse named reference %s: %w", value, err)
	}

	ref := OCIReference{Named: named}
	if ref.HasDigest() {
		err := ref.Digest().Validate()
		if err != nil {
			return OCIReference{}, fmt.Errorf("invalid digest for reference %s: %w", value, err)
		}
	}

	return ref, nil
}

// MustParseOCIReference parses the specified value as an OCIReference,
// panicking on any errors.
// Only use this for unit tests where you know the value is a reference.
func MustParseOCIReference(value string) OCIReference {
	ref, err := ParseOCIReference(value)
	if err != nil {
		panic(err)
	}
	return ref
}

var (
	_ json.Marshaler   = OCIReference{}
	_ json.Unmarshaler = &OCIReference{}
)

// OCIReference is a wrapper around a docker reference with convenience methods
// for decomposing and manipulating bundle references.
//
// It is designed to be safe to call even when uninitialized, returning empty
// strings when parts are requested that do not exist, such as calling Digest()
// when no digest is set on the reference.
type OCIReference struct {
	// Name is the wrapped reference that we are providing helper methods on top of
	Named reference.Named
}

func (r *OCIReference) UnmarshalJSON(bytes []byte) error {
	value := strings.TrimPrefix(strings.TrimSuffix(string(bytes), `"`), `"`)
	ref, err := ParseOCIReference(value)
	if err != nil {
		return err
	}
	r.Named = ref.Named
	return nil
}

func (r OCIReference) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, r.String())), nil
}

// Always print the original name provided, not
// the one with docker.io prefixed.
func (r OCIReference) String() string {
	if r.Named == nil {
		return ""
	}
	return reference.FamiliarString(r.Named)
}

// Repository portion of the reference.
// Example: docker.io/getporter/mybuns:v0.1.1 returns getporter/mybuns
func (r OCIReference) Repository() string {
	if r.Named == nil {
		return ""
	}
	return reference.FamiliarName(r.Named)
}

// Registry portion of the reference.
// Example: ghcr.io/getporter/mybuns:v0.1.1 returns ghcr.io
func (r OCIReference) Registry() string {
	if r.Named == nil {
		return ""
	}
	return reference.Domain(r.Named)
}

// IsRepositoryOnly determines if the reference is fully qualified
// with a tag/digest or if it only contains a repository.
// Example: ghcr.io/getporter/mybuns returns true
func (r OCIReference) IsRepositoryOnly() bool {
	return !r.HasTag() && !r.HasDigest()
}

// HasDigest determines if the reference has a digest portion.
// Example: ghcr.io/getporter/mybuns@sha256:abc123 returns true
func (r OCIReference) HasDigest() bool {
	if r.Named == nil {
		return false
	}

	_, ok := r.Named.(reference.Digested)
	return ok
}

// Digest portion of the reference.
// Example: ghcr.io/getporter/mybuns@sha256:abc123 returns sha256:abc123
func (r OCIReference) Digest() digest.Digest {
	if r.Named == nil {
		return ""
	}

	t, ok := r.Named.(reference.Digested)
	if ok {
		return t.Digest()
	}
	return ""
}

// HasTag determines if the reference has a tag portion.
// Example: ghcr.io/getporter/mybuns:latest returns true
func (r OCIReference) HasTag() bool {
	if r.Named == nil {
		return false
	}

	_, ok := r.Named.(reference.Tagged)
	return ok
}

// Tag portion of the reference.
// Example: ghcr.io/getporter/mybuns:latest returns latest
func (r OCIReference) Tag() string {
	if r.Named == nil {
		return ""
	}

	t, ok := r.Named.(reference.Tagged)
	if ok {
		return t.Tag()
	}
	return ""
}

// HasVersion detects if the reference tag is a bundle version (semver).
func (r OCIReference) HasVersion() bool {
	if r.Named == nil {
		return false
	}

	if tagged, ok := r.Named.(reference.Tagged); ok {
		_, err := semver.NewVersion(tagged.Tag())
		return err == nil
	}
	return false
}

// Version parses the reference tag as a bundle version (semver).
func (r OCIReference) Version() string {
	if r.Named == nil {
		return ""
	}

	if tagged, ok := r.Named.(reference.Tagged); ok {
		v, err := semver.NewVersion(tagged.Tag())
		if err == nil {
			return v.String()
		}
	}
	return ""
}

// WithVersion creates a new reference using the repository and the specified bundle version.
func (r OCIReference) WithVersion(version string) (OCIReference, error) {
	if r.Named == nil {
		return OCIReference{}, errors.New("OCIReference has not been initialized")
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return OCIReference{}, fmt.Errorf("invalid bundle version specified %s: %w", version, err)
	}

	newRef, err := reference.WithTag(r.Named, "v"+v.String())
	if err != nil {
		return OCIReference{}, err
	}
	return OCIReference{Named: newRef}, nil
}

// WithTag creates a new reference using the repository and the specified tag.
func (r OCIReference) WithTag(tag string) (OCIReference, error) {
	if r.Named == nil {
		return OCIReference{}, errors.New("OCIReference has not been initialized")
	}
	newRef, err := reference.WithTag(r.Named, tag)
	if err != nil {
		return OCIReference{}, err
	}
	return OCIReference{Named: newRef}, nil
}

// WithDigest creates a new reference using the repository and the specified digest.
func (r OCIReference) WithDigest(digest digest.Digest) (OCIReference, error) {
	if r.Named == nil {
		return OCIReference{}, errors.New("OCIReference has not been initialized")
	}
	newRef, err := reference.WithDigest(r.Named, digest)
	if err != nil {
		return OCIReference{}, err
	}
	return OCIReference{Named: newRef}, nil
}

// ParseRepositoryInfo returns additional metadata about the repository portion of the reference.
func (r OCIReference) ParseRepositoryInfo() (*registry.RepositoryInfo, error) {
	if r.Named == nil {
		return nil, errors.New("OCIReference has not been initialized")
	}
	return registry.ParseRepositoryInfo(r.Named)
}
