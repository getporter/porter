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
	"errors"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/pivotal/image-relocation/pkg/registry"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pivotal/image-relocation/pkg/image"
)

var (
	repoImageFunc      = remote.Image
	resolveFunc        = authn.DefaultKeychain.Resolve
	repoGetFunc        = remote.Get
	repoWriteFunc      = remote.Write
	repoIndexWriteFunc = remote.WriteIndex
)

func readRemoteImage(mfstWriter manifestWriter, idxWriter indexWriter) func(n image.Name) (registry.Image, error) {
	return func(n image.Name) (i registry.Image, e error) {
		auth, err := resolve(n)
		if err != nil {
			return nil, err
		}

		if n.Tag() == "" && n.Digest() == image.EmptyDigest {
			// use default tag
			n, err = n.WithTag("latest")
			if err != nil {
				return nil, err
			}
		}
		ref, err := name.ParseReference(n.String(), name.StrictValidation)
		if err != nil {
			return nil, err
		}

		desc, err := repoGetFunc(ref, remote.WithAuth(auth))
		if err != nil {
			return nil, err
		}

		switch desc.MediaType {
		case types.OCIImageIndex, types.DockerManifestList:
			idx, err := desc.ImageIndex()
			if err != nil {
				return nil, err
			}
			return newImageFromIndex(idx, n, idxWriter), nil
		default:
			// assume all other media types are images since some images don't set the media type
		}
		img, err := desc.Image()
		if err != nil {
			return nil, err
		}

		return newImageFromManifest(img, n, mfstWriter), nil
	}
}

func writeRemoteImage(i v1.Image, n image.Name) error {
	auth, err := resolve(n)
	if err != nil {
		return err
	}

	ref, err := getWriteReference(n)
	if err != nil {
		return err
	}

	return repoWriteFunc(ref, i, remote.WithAuth(auth), remote.WithTransport(http.DefaultTransport))
}

func writeRemoteIndex(i v1.ImageIndex, n image.Name) error {
	auth, err := resolve(n)
	if err != nil {
		return err
	}

	ref, err := getWriteReference(n)
	if err != nil {
		return err
	}

	return repoIndexWriteFunc(ref, i, remote.WithAuth(auth), remote.WithTransport(http.DefaultTransport))
}

func resolve(n image.Name) (authn.Authenticator, error) {
	if n == image.EmptyName {
		return nil, errors.New("empty image name invalid")
	}
	repo, err := name.NewRepository(n.WithoutTagOrDigest().String(), name.WeakValidation)
	if err != nil {
		return nil, err
	}

	return resolveFunc(repo.Registry)
}

func getWriteReference(n image.Name) (name.Reference, error) {
	// if target image reference is both tagged and digested, ignore the digest so the tag is preserved
	// (the digest will be preserved by go-containerregistry)
	if n.Tag() != "" {
		n = n.WithoutDigest()
	}

	return name.ParseReference(n.String(), name.WeakValidation)
}
