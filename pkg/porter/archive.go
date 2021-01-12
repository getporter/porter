package porter

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/imagestore"
	"github.com/cnabio/cnab-go/imagestore/construction"
	"github.com/docker/docker/pkg/archive"
	"github.com/pkg/errors"
)

// ArchiveOptions defines the valid options for performing an archive operation
type ArchiveOptions struct {
	BundleActionOptions
	ArchiveFile string
}

// Validate performs validation on the publish options
func (o *ArchiveOptions) Validate(args []string, p *Porter) error {
	if len(args) < 1 || args[0] == "" {
		return errors.New("destination file is required")
	}
	if len(args) > 1 {
		return errors.Errorf("only one positional argument may be specified, the archive file name, but multiple were received: %s", args)
	}
	o.ArchiveFile = args[0]

	if o.Reference == "" {
		return errors.New("must provide a value for --reference of the form REGISTRY/bundle:tag")
	}
	return o.BundleActionOptions.Validate(args, p)
}

// Archive is a composite function that generates a CNAB thick bundle. It will pull the invocation image, and
// any referenced images locally (if needed), export them to individual layers, generate a bundle.json and
// then generate a gzipped tar archive containing the bundle.json and the images
func (p *Porter) Archive(opts ArchiveOptions) error {
	dir := filepath.Dir(opts.ArchiveFile)
	if _, err := p.Config.FileSystem.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("parent directory %q does not exist", dir)
	}

	err := p.prepullBundleByReference(&opts.BundleActionOptions)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before building archive")
	}

	err = p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}
	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	bun, err := p.CNAB.LoadBundle(opts.CNABFile)
	if err != nil {
		return errors.Wrap(err, "couldn't open bundle for archiving")
	}
	// This allows you to export thin or thick bundles, we only support generating "thick" archives
	ctor, err := construction.NewConstructor(false)
	if err != nil {
		return err
	}

	dest, err := p.Config.FileSystem.OpenFile(opts.ArchiveFile, os.O_RDWR|os.O_CREATE, 0644)

	exp := &exporter{
		out:                   p.Config.Out,
		logs:                  p.Config.Out,
		bundle:                bun,
		destination:           dest,
		imageStoreConstructor: ctor,
	}
	if err := exp.export(); err != nil {
		return err
	}

	return nil
}

type exporter struct {
	out                   io.Writer
	logs                  io.Writer
	bundle                bundle.Bundle
	destination           io.Writer
	imageStoreConstructor imagestore.Constructor
	imageStore            imagestore.Store
}

func (ex *exporter) export() error {

	name := ex.bundle.Name + "-" + ex.bundle.Version
	archiveDir, err := ioutil.TempDir("", name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(archiveDir, 0644); err != nil {
		return err
	}
	defer os.RemoveAll(archiveDir)

	to, err := os.OpenFile(filepath.Join(archiveDir, "bundle.json"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer to.Close()
	_, err = ex.bundle.WriteTo(to)
	if err != nil {
		return errors.Wrap(err, "unable to write bundle.json in archive")
	}

	ex.imageStore, err = ex.imageStoreConstructor(imagestore.WithArchiveDir(archiveDir), imagestore.WithLogs(ex.logs))
	if err != nil {
		return fmt.Errorf("Error creating artifacts: %s", err)
	}

	if err := ex.prepareArtifacts(ex.bundle); err != nil {
		return fmt.Errorf("Error preparing artifacts: %s", err)
	}

	tarOptions := &archive.TarOptions{
		Compression:      archive.Gzip,
		IncludeFiles:     []string{"."},
		IncludeSourceDir: true,
	}
	rc, err := archive.TarWithOptions(archiveDir, tarOptions)
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.Copy(ex.destination, rc)
	return err
}

// prepareArtifacts pulls all images, verifies their digests and
// saves them to a directory called artifacts/ in the bundle directory
func (ex *exporter) prepareArtifacts(bun bundle.Bundle) error {
	for _, image := range bun.Images {
		if err := ex.addImage(image.BaseImage); err != nil {
			return err
		}
	}

	for _, in := range bun.InvocationImages {
		if err := ex.addImage(in.BaseImage); err != nil {
			return err
		}
	}

	return nil
}

// addImage pulls an image, adds it to the artifacts/ directory, and verifies its digest
func (ex *exporter) addImage(image bundle.BaseImage) error {
	dig, err := ex.imageStore.Add(image.Image)
	if err != nil {
		return err
	}
	return checkDigest(image, dig)
}

// checkDigest compares the content digest of the given image to the given content digest and returns an error if they
// are both non-empty and do not match
func checkDigest(image bundle.BaseImage, dig string) error {
	digestFromManifest := image.Digest
	if dig == "" || digestFromManifest == "" {
		return nil
	}
	if digestFromManifest != dig {
		return fmt.Errorf("content digest mismatch: image %s has digest %s but the digest should be %s according to the bundle manifest", image.Image, dig, digestFromManifest)
	}
	return nil
}
