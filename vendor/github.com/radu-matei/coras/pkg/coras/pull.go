package coras

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
)

// Pull pulls a CNAB bundle file from an OCI registry
func Pull(targetRef, outputBundle string, exported bool) error {
	var pulledFile = CNABThinBundleFileName
	var mediaType = CNABThinMediaType
	if exported {
		pulledFile = CNABThickBundleFileName
		mediaType = CNABThickMediaType
	}

	fsRoot, err := ioutil.TempDir("", "coras-")
	if err != nil {
		return fmt.Errorf("cannot create temporary directory: %v", err)
	}
	defer os.RemoveAll(fsRoot)

	fs := content.NewFileStore(fsRoot)
	defer fs.Close()

	_, _, err = oras.Pull(context.Background(), newResolver(), targetRef, fs, oras.WithAllowedMediaTypes([]string{mediaType}))
	if err != nil {
		return fmt.Errorf("cannot pull: %v", err)
	}

	input, err := ioutil.ReadFile(path.Join(fsRoot, pulledFile))
	if err != nil {
		return err
	}

	return ioutil.WriteFile(outputBundle, input, 0644)
}
