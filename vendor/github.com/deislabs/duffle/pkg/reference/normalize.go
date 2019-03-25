package reference

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/distribution/digestset"
	digest "github.com/opencontainers/go-digest"
)

const (
	defaultTag = "latest"
)

// normalizedNamed represents a name which has been
// normalized and has a familiar form. A familiar name
// is what is used in Duffle's UI. An example normalized
// name is "hub.cnlabs.io/library/ubuntu" and corresponding
// familiar name of "library/ubuntu".
type normalizedNamed interface {
	Named
	Familiar() Named
}

// ParseNormalizedNamed parses a string into a named reference
// transforming a familiar name from Docker UI to a fully
// qualified reference. If the value may be an identifier
// use ParseAnyReference.
func ParseNormalizedNamed(s string) (Named, error) {
	if ok := anchoredIdentifierRegexp.MatchString(s); ok {
		return nil, fmt.Errorf("repository name (%s) cannot be a 64-byte hexadecimal strings", s)
	}
	domain, remainder := splitDockerDomain(s)
	var remoteName string
	if tagSep := strings.IndexRune(remainder, ':'); tagSep > -1 {
		remoteName = remainder[:tagSep]
	} else {
		remoteName = remainder
	}
	if strings.ToLower(remoteName) != remoteName {
		return nil, errors.New("in a reference name, the repository part must be lowercase")
	}

	rem := domain + "/" + remainder
	if domain == "" {
		rem = remainder
	}
	ref, err := Parse(rem)
	if err != nil {
		return nil, err
	}
	named, isNamed := ref.(Named)
	if !isNamed {
		return nil, fmt.Errorf("reference %s has no name", ref.String())
	}
	return named, nil
}

// splitDockerDomain splits a repository name to domain and remotename string.
// If no valid domain is found, the default domain is used. Repository name
// needs to be already validated before.
func splitDockerDomain(name string) (domain, remainder string) {
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost") {
		domain, remainder = "", name
	} else {
		domain, remainder = name[:i], name[i+1:]
	}
	return
}

// TagNameOnly adds the default tag "latest" to a reference if it only has
// a repo name.
func TagNameOnly(ref Named) Named {
	if IsNameOnly(ref) {
		namedTagged, err := WithTag(ref, defaultTag)
		if err != nil {
			// Default tag must be valid, to create a NamedTagged
			// type with non-validated input the WithTag function
			// should be used instead
			panic(err)
		}
		return namedTagged
	}
	return ref
}

// ParseAnyReference parses a reference string as a possible identifier,
// full digest, or familiar name.
func ParseAnyReference(ref string) (Reference, error) {
	if ok := anchoredIdentifierRegexp.MatchString(ref); ok {
		return digestReference("sha256:" + ref), nil
	}
	if dgst, err := digest.Parse(ref); err == nil {
		return digestReference(dgst), nil
	}

	return ParseNormalizedNamed(ref)
}

// ParseAnyReferenceWithSet parses a reference string as a possible short
// identifier to be matched in a digest set, a full digest, or familiar name.
func ParseAnyReferenceWithSet(ref string, ds *digestset.Set) (Reference, error) {
	if ok := anchoredShortIdentifierRegexp.MatchString(ref); ok {
		dgst, err := ds.Lookup(ref)
		if err == nil {
			return digestReference(dgst), nil
		}
	} else {
		if dgst, err := digest.Parse(ref); err == nil {
			return digestReference(dgst), nil
		}
	}

	return ParseNormalizedNamed(ref)
}
