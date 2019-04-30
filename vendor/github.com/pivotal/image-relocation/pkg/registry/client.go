/*
 * Copyright (c) 2019-Present Pivotal Software, Inc. All rights reserved.
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

package registry

import (
	"fmt"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/image-relocation/pkg/image"
)

// Client provides a way of interacting with image registries.
type Client interface {
	// Digest returns the digest of the given image or an error if the image does not exist or the digest is unavailable.
	Digest(image.Name) (image.Digest, error)

	// Copy copies the given source image to the given target and returns the image's digest (which is preserved).
	Copy(source image.Name, target image.Name) (image.Digest, error)

	// NewLayout creates a Layout for the Client and creates a corresponding directory containing a new OCI image layout at
	// the given file system path.
	NewLayout(path string) (Layout, error)

	// ReadLayout creates a Layout for the Client from the given file system path of a directory containing an existing
	// OCI image layout.
	ReadLayout(path string) (Layout, error)
}

type client struct {
	readRemoteImage  func(n image.Name) (v1.Image, error)
	writeRemoteImage func(i v1.Image, n image.Name) error
}

// NewRegistryClient returns a new Client.
func NewRegistryClient() Client {
	return &client{
		readRemoteImage:  readRemoteImage,
		writeRemoteImage: writeRemoteImage,
	}
}

func (r *client) Digest(n image.Name) (image.Digest, error) {
	img, err := r.readRemoteImage(n)
	if err != nil {
		return image.EmptyDigest, err
	}

	hash, err := img.Digest()
	if err != nil {
		return image.EmptyDigest, err
	}

	return image.NewDigest(hash.String())
}

func (r *client) Copy(source image.Name, target image.Name) (image.Digest, error) {
	img, err := r.readRemoteImage(source)
	if err != nil {
		return image.EmptyDigest, fmt.Errorf("failed to read image %v: %v", source, err)
	}

	hash, err := img.Digest()
	if err != nil {
		return image.EmptyDigest, fmt.Errorf("failed to read digest of image %v: %v", source, err)
	}

	err = r.writeRemoteImage(img, target)
	if err != nil {
		return image.EmptyDigest, fmt.Errorf("failed to write image %v: %v", target, err)
	}

	return image.NewDigest(hash.String())
}
