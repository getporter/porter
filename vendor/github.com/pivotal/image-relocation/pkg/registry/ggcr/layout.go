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

	"github.com/pivotal/image-relocation/pkg/registry"
	"github.com/pivotal/image-relocation/pkg/registry/ggcr/path"

	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"

	"github.com/pivotal/image-relocation/pkg/image"
)

const refNameAnnotation = "org.opencontainers.image.ref.name"

type imageLayout struct {
	registryClient RegistryClient
	layoutPath     path.LayoutPath
}

func NewImageLayout(registryClient RegistryClient, layoutPath path.LayoutPath) registry.Layout {
	return &imageLayout{
		registryClient: registryClient,
		layoutPath:     layoutPath,
	}
}

// appendable is an interface implemented by types which can be appended to an OCI image layout.
type appendable interface {
	// appendToLayout appends the image to a given OCI image layout using the given layout options.
	appendToLayout(layoutPath path.LayoutPath, options ...layout.Option) error
}

func (l *imageLayout) Add(n image.Name) (image.Digest, error) {
	img, err := l.registryClient.ReadRemoteImage(n)
	if err != nil {
		return image.EmptyDigest, err
	}

	annotations := map[string]string{
		refNameAnnotation: n.String(),
	}
	if img, ok := img.(appendable); ok {
		if err := img.appendToLayout(l.layoutPath, layout.WithAnnotations(annotations)); err != nil {
			return image.EmptyDigest, err
		}
	}

	hash, err := img.Digest()
	if err != nil {
		return image.EmptyDigest, err
	}

	return image.NewDigest(hash.String())
}

func (l *imageLayout) Push(digest image.Digest, n image.Name) error {
	img, err := l.findByDigest(digest)
	if err != nil {
		return err
	}

	_, _, err = img.Write(n)
	if err != nil {
		return fmt.Errorf("failed to write image %v to %v: %v", digest, n, err)
	}

	return nil
}

func (l *imageLayout) Find(n image.Name) (image.Digest, error) {
	imageIndex, err := l.layoutPath.ImageIndex()
	if err != nil {
		return image.EmptyDigest, err
	}
	indexMan, err := imageIndex.IndexManifest()
	if err != nil {
		return image.EmptyDigest, err
	}

	for _, imageMan := range indexMan.Manifests {
		if ref, ok := imageMan.Annotations[refNameAnnotation]; ok {
			r, err := image.NewName(ref)
			if err != nil {
				return image.EmptyDigest, err
			}
			if r == n {
				return image.NewDigest(imageMan.Digest.String())
			}
		}
	}

	return image.EmptyDigest, fmt.Errorf("image %v not found in layout", n)
}

func (l *imageLayout) findByDigest(digest image.Digest) (registry.Image, error) {
	hash, err := v1.NewHash(digest.String())
	if err != nil {
		return nil, err
	}
	imageIndex, err := l.layoutPath.ImageIndex()
	if err != nil {
		return nil, err
	}

	im, err := imageIndex.Image(hash)
	if err == nil {
		return l.registryClient.NewImageFromManifest(im), nil
	}

	idx, err := imageIndex.ImageIndex(hash)
	if err == nil {
		return l.registryClient.NewImageFromIndex(idx), nil
	}

	return nil, err
}
