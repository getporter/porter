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
	"os"

	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/pivotal/image-relocation/pkg/image"
)

const (
	outputDirPermissions = 0755
	refNameAnnotation    = "org.opencontainers.image.ref.name"
)

func (r *client) NewLayout(path string) (registry.Layout, error) {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err := os.MkdirAll(path, outputDirPermissions); err != nil {
			return nil, err
		}
	}

	lp, err := layout.Write(path, empty.Index)
	if err != nil {
		return nil, err
	}

	return &imageLayout{
		registryClient: r,
		layoutPath:     lp,
	}, nil
}

func (r *client) ReadLayout(path string) (registry.Layout, error) {
	lp, err := layout.FromPath(path)
	if err != nil {
		return nil, err
	}
	return &imageLayout{
		registryClient: r,
		layoutPath:     lp,
	}, nil
}

type imageLayout struct {
	registryClient *client
	layoutPath     registry.LayoutPath
}

func NewImageLayout(registryClient *client, layoutPath registry.LayoutPath) registry.Layout {
	return &imageLayout{
		registryClient: registryClient,
		layoutPath:     layoutPath,
	}
}

func (l *imageLayout) Add(n image.Name) (image.Digest, error) {
	img, err := l.registryClient.readRemoteImage(n)
	if err != nil {
		return image.EmptyDigest, err
	}

	annotations := map[string]string{
		refNameAnnotation: n.String(),
	}
	if err:= img.AppendToLayout(l.layoutPath, layout.WithAnnotations(annotations)); err != nil {
		return image.EmptyDigest, err
	}

	hash, err := img.Digest()
	if err != nil {
		return image.EmptyDigest, err
	}

	return image.NewDigest(hash.String())
}

func (l *imageLayout) Push(digest image.Digest, n image.Name) error {
	hash, err := v1.NewHash(digest.String())
	if err != nil {
		return err
	}
	imageIndex, err := l.layoutPath.ImageIndex()
	if err != nil {
		return err
	}
	i, err := imageIndex.Image(hash)
	if err != nil {
		return err
	}

	return l.registryClient.writeRemoteImage(i, n)
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
