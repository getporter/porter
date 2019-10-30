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
	"github.com/pivotal/image-relocation/pkg/image"
)

// A Layout abstracts an OCI image layout on disk.
type Layout interface {
	// Add adds the image at the given image reference to the layout and returns the image's digest.
	Add(n image.Name) (image.Digest, error)

	// Push pushes the image with the given digest from the layout to the given image reference.
	Push(digest image.Digest, name image.Name) error

	// Find returns the digest of an image in the layout with the given image reference.
	Find(n image.Name) (image.Digest, error)
}

