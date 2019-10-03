package ocilayout

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pivotal/image-relocation/pkg/image"
	"github.com/pivotal/image-relocation/pkg/registry"
	"github.com/pivotal/image-relocation/pkg/registry/ggcr"

	"github.com/deislabs/cnab-go/imagestore"
)

// ociLayout is an image store which stores images as an OCI image layout.
type ociLayout struct {
	layout registry.Layout
	logs   io.Writer
}

func Create(options ...imagestore.Option) (imagestore.Store, error) {
	parms := imagestore.Create(options...)

	layoutDir := filepath.Join(parms.ArchiveDir, "artifacts", "layout")
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		return nil, err
	}

	layout, err := ggcr.NewRegistryClient().NewLayout(layoutDir)
	if err != nil {
		return nil, err
	}

	return &ociLayout{
		layout: layout,
		logs:   parms.Logs,
	}, nil
}

func LocateOciLayout(archiveDir string) (imagestore.Store, error) {
	layoutDir := filepath.Join(archiveDir, "artifacts", "layout")
	if _, err := os.Stat(layoutDir); os.IsNotExist(err) {
		return nil, err
	}
	layout, err := ggcr.NewRegistryClient().ReadLayout(layoutDir)
	if err != nil {
		return nil, err
	}

	return &ociLayout{
		layout: layout,
		logs:   ioutil.Discard,
	}, nil
}

func (o *ociLayout) Add(im string) (string, error) {
	n, err := image.NewName(im)
	if err != nil {
		return "", err
	}

	dig, err := o.layout.Add(n)
	if err != nil {
		return "", err
	}

	return dig.String(), nil
}

func (o *ociLayout) Push(dig image.Digest, src image.Name, dst image.Name) error {
	if dig == image.EmptyDigest {
		var err error
		dig, err = o.layout.Find(src)
		if err != nil {
			return err
		}
	}
	return o.layout.Push(dig, dst)
}
