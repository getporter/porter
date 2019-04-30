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

import "github.com/opencontainers/go-digest"

// Digest provides a CAS address of an image.
type Digest struct {
	dig digest.Digest
}

// NewDigest returns the Digest for a given digest string or an error if the digest string is invalid.
// The digest string must be of the form "alg:hash" where alg is an algorithm, such as sha256, and hash
// is a string output by that algorithm.
func NewDigest(dig string) (Digest, error) {
	d, err := digest.Parse(dig)
	if err != nil {
		return EmptyDigest, err
	}
	return Digest{d}, nil
}

// EmptyDigest is an invalid, zero value for Digest.
var EmptyDigest Digest

func init() {
	EmptyDigest = Digest{""}
}

// String returns the string form of the Digest.
func (d Digest) String() string {
	return string(d.dig)
}
