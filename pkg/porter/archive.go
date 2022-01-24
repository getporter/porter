package porter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/carolynvs/aferox"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/imagestore"
	"github.com/cnabio/cnab-go/imagestore/construction"
	"github.com/mholt/archiver/v4"
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
func (p *Porter) Archive(ctx context.Context, opts ArchiveOptions) error {
	dir := filepath.Dir(opts.ArchiveFile)
	if _, err := p.Config.FileSystem.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("parent directory %q does not exist", dir)
	}

	bundleRef, err := p.resolveBundleReference(ctx, &opts.BundleActionOptions)
	if err != nil {
		return err
	}

	// This allows you to export thin or thick bundles, we only support generating "thick" archives
	ctor, err := construction.NewConstructor(false)
	if err != nil {
		return err
	}

	dest, err := p.Config.FileSystem.OpenFile(opts.ArchiveFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return errors.Wrapf(err, "couldn't create archive file %s", opts.ArchiveFile)
	}
	defer dest.Close()

	exp := &exporter{
		fs:                    p.Config.FileSystem,
		out:                   p.Config.Out,
		logs:                  p.Config.Out,
		bundle:                bundleRef.Definition,
		destination:           dest,
		imageStoreConstructor: ctor,
	}
	if err := exp.export(); err != nil {
		return err
	}

	return nil
}

type exporter struct {
	fs                    aferox.Aferox
	out                   io.Writer
	logs                  io.Writer
	bundle                cnab.ExtendedBundle
	destination           io.Writer
	imageStoreConstructor imagestore.Constructor
	imageStore            imagestore.Store
}

func (ex *exporter) export() error {

	name := ex.bundle.Name + "-" + ex.bundle.Version
	archiveDir, err := ex.fs.TempDir("", name)
	if err != nil {
		return err
	}
	if err := ex.fs.MkdirAll(archiveDir, 0644); err != nil {
		return err
	}
	defer ex.fs.RemoveAll(archiveDir)

	to, err := ex.fs.OpenFile(filepath.Join(archiveDir, "bundle.json"), os.O_RDWR|os.O_CREATE, 0600)
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
		return fmt.Errorf("error creating artifacts: %s", err)
	}

	if err := ex.prepareArtifacts(ex.bundle); err != nil {
		return fmt.Errorf("error preparing artifacts: %s", err)
	}

	return ex.compressArchive(archiveDir)
}

func (ex *exporter) compressArchive(src string) error {
	// Archive the items inside the source directory without including it as a top-level directory
	childItems, err := ex.fs.ReadDir(src)
	if err != nil {
		return errors.Wrapf(err, "error reading directory %s", src)
	}

	files := make(map[string]string, len(childItems))
	for _, child := range childItems {
		srcFile := filepath.Join(src, child.Name())
		files[srcFile] = child.Name()
	}

	arFiles, err := archiver.FilesFromDisk(&archiver.FromDiskOptions{ClearAttributes: true}, files)
	if err != nil {
		return err
	}
	format := archiver.CompressedArchive{
		Compression: archiver.Gz{},
		Archival:    archiver.Tar{},
	}

	// create the archive
	err = format.Archive(context.Background(), ex.destination, arFiles)
	return errors.Wrapf(err, "error archiving the bundle")
}

// prepareArtifacts pulls all images, verifies their digests and
// saves them to a directory called artifacts/ in the bundle directory
func (ex *exporter) prepareArtifacts(bun cnab.ExtendedBundle) error {
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
