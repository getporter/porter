package cnab

import (
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/opencontainers/go-digest"
)

type BundleReference struct {
	Reference     OCIReference
	Digest        digest.Digest
	Definition    ExtendedBundle
	RelocationMap relocation.ImageRelocationMap
}
