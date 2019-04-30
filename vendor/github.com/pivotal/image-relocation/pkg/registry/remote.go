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
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/pivotal/image-relocation/pkg/image"
)

func readRemoteImage(n image.Name) (v1.Image, error) {
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

	return remote.Image(ref, remote.WithAuth(auth))
}

func writeRemoteImage(i v1.Image, n image.Name) error {
	auth, err := resolve(n)
	if err != nil {
		return err
	}

	ref, err := name.ParseReference(n.String(), name.WeakValidation)
	if err != nil {
		return err
	}

	return remote.Write(ref, i, auth, http.DefaultTransport)
}

func resolve(n image.Name) (authn.Authenticator, error) {
	repo, err := name.NewRepository(n.WithoutTagOrDigest().String(), name.WeakValidation)
	if err != nil {
		return nil, err
	}

	return authn.DefaultKeychain.Resolve(repo.Registry)
}
