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
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
)

func newImageFromManifest(img v1.Image, nm image.Name, mfstWriter manifestWriter) *imageManifest {
	return &imageManifest{manifest: img, nm: nm, mfstWriter: mfstWriter}
}

type imageManifest struct {
	manifest   v1.Image
	nm         image.Name
	mfstWriter manifestWriter
}

func (m *imageManifest) Digest() (image.Digest, error) {
	hash, err := m.manifest.Digest()
	if err != nil {
		return image.EmptyDigest, err
	}
	return image.NewDigest(hash.String())
}

func (m *imageManifest) Write(target image.Name) (image.Digest, int64, error) {
	dig, err := m.Digest()
	if err != nil {
		return image.EmptyDigest, 0, fmt.Errorf("failed to read digest of image %v: %v", m.nm, err)
	}

	err = m.mfstWriter(m.manifest, target)
	if err != nil {
		return image.EmptyDigest, 0, fmt.Errorf("failed to write image %v: %v", target, err)
	}

	rawManifest, err := m.manifest.RawManifest()
	if err != nil {
		return image.EmptyDigest, 0, fmt.Errorf("failed to get raw manifest of image %v: %v", m.nm, err)
	}

	return dig, int64(len(rawManifest)), nil
}

func (m *imageManifest) AppendToLayout(layoutPath registry.LayoutPath, options ...layout.Option) error {
	return layoutPath.AppendImage(m.manifest, options...)
}

type imageIndex struct {
	index     v1.ImageIndex
	nm        image.Name
	idxWriter indexWriter
}

func (i *imageIndex) Digest() (image.Digest, error) {
	hash, err := i.index.Digest()
	if err != nil {
		return image.EmptyDigest, err
	}
	return image.NewDigest(hash.String())
}

func (i *imageIndex) Write(target image.Name) (image.Digest, int64, error) {
	dig, err := i.Digest()
	if err != nil {
		return image.EmptyDigest, 0, fmt.Errorf("failed to read digest of image index %v: %v", i.nm, err)
	}

	err = i.idxWriter(i.index, target)
	if err != nil {
		return image.EmptyDigest, 0, fmt.Errorf("failed to write image index %v: %v", target, err)
	}

	rawManifest, err := i.index.RawManifest()
	if err != nil {
		return image.EmptyDigest, 0, fmt.Errorf("failed to get raw manifest of image index %v: %v", i.nm, err)
	}

	return dig, int64(len(rawManifest)), nil
}

func (i *imageIndex) AppendToLayout(layoutPath registry.LayoutPath, options ...layout.Option) error {
	return layoutPath.AppendIndex(i.index, options...)
}

func newImageFromIndex(idx v1.ImageIndex, nm image.Name, idxWriter indexWriter) *imageIndex {
	return &imageIndex{index: idx, nm: nm, idxWriter: idxWriter}
}
