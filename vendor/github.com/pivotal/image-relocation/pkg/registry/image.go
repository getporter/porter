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
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/pivotal/image-relocation/pkg/image"
)

// Image represents an abstract image which could be an image manifest or an image index (e.g. a multi-arch image).
type Image interface {
	// Digest returns the repository digest of the image.
	Digest() (image.Digest, error)

	// Write writes the image to a given reference and returns the image's digest and size.
	Write(target image.Name) (image.Digest, int64, error)

	// AppendToLayout appends the image to a given OCI image layout using the given layout options.
	AppendToLayout(layoutPath LayoutPath, options ...layout.Option) error
}
