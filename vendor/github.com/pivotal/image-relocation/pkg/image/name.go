/*
 * Copyright (c) 2018-Present Pivotal Software, Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package image

import (
	"fmt"
	"path"
	"strings"

	"github.com/docker/distribution/reference"
)

const (
	dockerHubHost     = "docker.io"
	fullDockerHubHost = "index.docker.io"
)

// Name is a named image reference. It can refer to an image manifest or a manifest list (e.g. a multi-arch image).
type Name struct {
	ref reference.Named
}

// EmptyName is an invalid, zero value for Name.
var EmptyName Name

func init() {
	EmptyName = Name{nil}
}

// NewName returns the Name for the given image reference or an error if the image reference is invalid.
func NewName(i string) (Name, error) {
	ref, err := reference.ParseNormalizedNamed(i)
	if err != nil {
		return Name{}, fmt.Errorf("invalid image reference: %q", i)
	}
	return Name{ref}, nil
}

// Normalize returns a fully-qualified equivalent to the Name. Useful on synonyms.
func (img Name) Normalize() Name {
	if img.ref == nil {
		return EmptyName
	}
	ref, err := NewName(img.String())
	if err != nil {
		panic(err) // should never happen
	}
	return ref
}

// Name returns the string form of the Name without any tag or digest.
func (img Name) Name() string {
	return img.ref.Name()
}

// String returns a string representation of the Name.
func (img Name) String() string {
	if img.ref == nil {
		return ""
	}
	return img.ref.String()
}

// Host returns the host of the Name. See also Path.
func (img Name) Host() string {
	h, _ := img.parseHostPath()
	return h
}

// Path returns the path of the name. See also Host.
func (img Name) Path() string {
	_, p := img.parseHostPath()
	return p
}

// Tag returns the tag of the Name or an empty string if the Name is not tagged.
func (img Name) Tag() string {
	if taggedRef, ok := img.ref.(reference.Tagged); ok {
		return taggedRef.Tag()
	}
	return ""
}

// WithTag returns a new Name with the same value as the Name, but with the given tag. It returns an error if and only
// if the tag is invalid.
func (img Name) WithTag(tag string) (Name, error) {
	namedTagged, err := reference.WithTag(img.ref, tag)
	if err != nil {
		return EmptyName, fmt.Errorf("Cannot apply tag %s to image.Name %v: %v", tag, img, err)
	}
	return Name{namedTagged}, nil
}

// Digest returns the digest of the Name or EmptyDigest if the Name does not have a digest.
func (img Name) Digest() Digest {
	if digestedRef, ok := img.ref.(reference.Digested); ok {
		d, err := NewDigest(string(digestedRef.Digest()))
		if err != nil {
			panic(err) // should never happen
		}
		return d
	}
	return EmptyDigest
}

// WithoutTagOrDigest returns a new Name with the same value as the Name, but with any tag or digest removed.
func (img Name) WithoutTagOrDigest() Name {
	return Name{reference.TrimNamed(img)}
}

// WithDigest returns a new Name with the same value as the Name, but with the given digest. It returns an error if and only
// if the digest is invalid.
func (img Name) WithDigest(digest Digest) (Name, error) {
	digested, err := reference.WithDigest(img.ref, digest.dig)
	if err != nil {
		return EmptyName, fmt.Errorf("Cannot apply digest %s to image.Name %v: %v", digest, img, err)
	}

	return Name{digested}, nil
}

// WithoutDigest returns a new Name with the same value as the Name, but with any digest removed. It preserves any tag.
func (img Name) WithoutDigest() Name {
	n := img.WithoutTagOrDigest()
	tag := img.Tag()
	if tag == "" {
		return n
	}
	n, _ = n.WithTag(tag)
	return n
}

// Synonyms returns the image names equivalent to a given image name. A synonym is not necessarily
// normalized: in particular it may not have a host name.
func (img Name) Synonyms() []Name {
	if img.ref == nil {
		return []Name{EmptyName}
	}
	imgHost, imgRepoPath := img.parseHostPath()
	nameMap := map[Name]struct{}{img: {}}

	if imgHost == dockerHubHost {
		elidedImg := imgRepoPath
		name, err := synonym(img, elidedImg)
		if err == nil {
			nameMap[name] = struct{}{}
		}

		elidedImgElements := strings.Split(elidedImg, "/")
		if len(elidedImgElements) == 2 && elidedImgElements[0] == "library" {
			name, err := synonym(img, elidedImgElements[1])
			if err == nil {
				nameMap[name] = struct{}{}
			}
		}

		fullImg := path.Join(fullDockerHubHost, imgRepoPath)
		name, err = synonym(img, fullImg)
		if err == nil {
			nameMap[name] = struct{}{}
		}

		dockerImg := path.Join(dockerHubHost, imgRepoPath)
		name, err = synonym(img, dockerImg)
		if err == nil {
			nameMap[name] = struct{}{}
		}
	}

	names := []Name{}
	for n := range nameMap {
		names = append(names, n)
	}

	return names
}

func synonym(original Name, newName string) (Name, error) {
	named, err := reference.WithName(newName)
	if err != nil {
		return EmptyName, err
	}

	if taggedRef, ok := original.ref.(reference.Tagged); ok {
		named, err = reference.WithTag(named, taggedRef.Tag())
		if err != nil {
			return EmptyName, err
		}
	}

	if digestedRef, ok := original.ref.(reference.Digested); ok {
		named, err = reference.WithDigest(named, digestedRef.Digest())
		if err != nil {
			return EmptyName, err
		}
	}

	return Name{named}, nil
}

func (img Name) parseHostPath() (host string, repoPath string) {
	s := strings.SplitN(img.ref.Name(), "/", 2)
	if len(s) == 1 {
		return img.Normalize().parseHostPath()
	}
	return s[0], s[1]
}
