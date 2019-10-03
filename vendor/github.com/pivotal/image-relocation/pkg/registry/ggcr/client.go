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

package ggcr

import (
	"fmt"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
)

type manifestWriter func(i v1.Image, n image.Name) error
type indexWriter func(i v1.ImageIndex, n image.Name) error

type client struct {
	readRemoteImage  func(n image.Name) (registry.Image, error)
	writeRemoteImage manifestWriter
	writeRemoteIndex indexWriter
}

// NewRegistryClient returns a new Client.
func NewRegistryClient() registry.Client {
	return &client{
		readRemoteImage:  readRemoteImage(writeRemoteImage, writeRemoteIndex),
		writeRemoteImage: writeRemoteImage,
		writeRemoteIndex: writeRemoteIndex,
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

func (r *client) Copy(source image.Name, target image.Name) (image.Digest, int64, error) {
	img, err := r.readRemoteImage(source)
	if err != nil {
		return image.EmptyDigest, 0, fmt.Errorf("failed to read image %v: %v", source, err)
	}

	sourceDigest, err := img.Digest()
	if err != nil {
		return image.EmptyDigest, 0, fmt.Errorf("failed to read digest of image %v: %v", source, err)
	}

	targetDigest, s, err := img.Write(target)
	if err != nil {
		return image.EmptyDigest, 0 , fmt.Errorf("failed to write image %v: %v", target, err)
	}
	if sourceDigest != targetDigest {
		return image.EmptyDigest, 0, fmt.Errorf("failed to preserve digest of image %v: source digest %v, target digest %v", source, sourceDigest, targetDigest)
	}
	return targetDigest, s, err
}
