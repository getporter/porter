package cnabtooci

import (
	"get.porter.sh/porter/pkg/cnab"
)

// BundleMetadata represents summary information about a bundle in a registry.
type BundleMetadata struct {
	cnab.BundleReference
}
