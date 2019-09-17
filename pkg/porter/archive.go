package porter

import (
	"fmt"

	"github.com/deislabs/porter/pkg/context"
)

// ArchiveOptions defines the valid options for performing an archive operation
type ArchiveOptions struct {
	bundleFileOptions
	File string
}

// Validate performs validation on the publish options
func (o *ArchiveOptions) Validate(args []string, cxt *context.Context) error {
	return nil
}

// Archive is a composite function that generates a CNAB thick bundle. It will pull the invocation image, and
// any referenced images locally (if needed), export them to individual layers, generate a bundle.json and
// then generate a gzipped tar archive containing the bundle.json and the images
func (p *Porter) Archive(opts ArchiveOptions) error {
	return fmt.Errorf("archive is not yet implemented")
}
