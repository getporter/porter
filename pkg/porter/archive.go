package porter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/carolynvs/aferox"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/imagestore"
	"github.com/cnabio/cnab-go/imagestore/construction"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/docker/docker/pkg/archive"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// ArchiveOptions defines the valid options for performing an archive operation
type ArchiveOptions struct {
	BundleActionOptions
	ArchiveFile string
}

// Validate performs validation on the publish options
func (o *ArchiveOptions) Validate(ctx context.Context, args []string, p *Porter) error {
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
	return o.BundleActionOptions.Validate(ctx, args, p)
}

// Archive is a composite function that generates a CNAB thick bundle. It will pull the invocation image, and
// any referenced images locally (if needed), export them to individual layers, generate a bundle.json and
// then generate a gzipped tar archive containing the bundle.json and the images
func (p *Porter) Archive(ctx context.Context, opts ArchiveOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	dir := filepath.Dir(opts.ArchiveFile)
	if _, err := p.Config.FileSystem.Stat(dir); os.IsNotExist(err) {
		return log.Error(fmt.Errorf("parent directory %q does not exist", dir))
	}

	bundleRef, err := p.resolveBundleReference(ctx, &opts.BundleActionOptions)
	if err != nil {
		return log.Error(err)
	}

	// This allows you to export thin or thick bundles, we only support generating "thick" archives
	ctor, err := construction.NewConstructor(false)
	if err != nil {
		return log.Error(err)
	}

	dest, err := p.Config.FileSystem.OpenFile(opts.ArchiveFile, os.O_RDWR|os.O_CREATE, pkg.FileModeWritable)
	if err != nil {
		return log.Error(err)
	}

	exp := &exporter{
		fs:                    p.Config.FileSystem,
		out:                   p.Config.Out,
		logs:                  p.Config.Out,
		bundle:                bundleRef.Definition,
		relocationMap:         bundleRef.RelocationMap,
		destination:           dest,
		imageStoreConstructor: ctor,
	}
	if err := exp.export(); err != nil {
		return log.Error(err)
	}

	return nil
}

type exporter struct {
	fs                    aferox.Aferox
	out                   io.Writer
	logs                  io.Writer
	bundle                cnab.ExtendedBundle
	relocationMap         relocation.ImageRelocationMap
	destination           io.Writer
	imageStoreConstructor imagestore.Constructor
	imageStore            imagestore.Store
}

func (ex *exporter) export() error {
	name := ex.bundle.Name + "-" + ex.bundle.Version
	archiveDir, err := ex.createArchiveFolder(name)
	if err != nil {
		return fmt.Errorf("can not create archive folder: %w", err)
	}
	defer ex.fs.RemoveAll(archiveDir)

	bundleFile, err := ex.fs.OpenFile(filepath.Join(archiveDir, "bundle.json"), os.O_RDWR|os.O_CREATE, pkg.FileModeWritable)
	if err != nil {
		return err
	}
	defer bundleFile.Close()
	_, err = ex.bundle.WriteTo(bundleFile)
	if err != nil {
		return errors.Wrap(err, "unable to write bundle.json in archive")
	}

	reloData, err := json.Marshal(ex.relocationMap)
	if err != nil {
		return err
	}
	err = ex.fs.WriteFile(filepath.Join(archiveDir, "relocation-mapping.json"), reloData, pkg.FileModeWritable)
	if err != nil {
		return errors.Wrap(err, "unable to write relocation-mapping.json in archive")
	}

	ex.imageStore, err = ex.imageStoreConstructor(imagestore.WithArchiveDir(archiveDir), imagestore.WithLogs(ex.logs))
	if err != nil {
		return fmt.Errorf("error creating artifacts: %s", err)
	}

	if err := ex.prepareArtifacts(ex.bundle); err != nil {
		return fmt.Errorf("error preparing artifacts: %s", err)
	}

	if err := ex.chtimes(archiveDir); err != nil {
		return fmt.Errorf("error preparing artifacts: %s", err)
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

// chtimes updates all paths under the provided archive path with a constant
// atime and mtime (Unix time 0), such that the shasum of the resulting archive
// will not change between repeated archival executions using the same bundle.
// See: https://unix.stackexchange.com/questions/346789/compressing-two-identical-folders-give-different-result
func (ex *exporter) chtimes(path string) error {
	err := filepath.Walk(path,
		func(subpath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			err = ex.fs.Chtimes(subpath, time.Unix(0, 0), time.Unix(0, 0))
			if err != nil {
				return err
			}
			return nil
		})
	if err != nil {
		return err
	}
	return nil
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

// addImage pulls an image using relocation map, adds it to the artifacts/ directory, and verifies its digest
func (ex *exporter) addImage(base bundle.BaseImage) error {
	if ex.relocationMap == nil {
		return errors.New("relocation map is not provided")
	}
	location, ok := ex.relocationMap[base.Image]
	if !ok {
		return fmt.Errorf("can not locate the referenced image: %s", base.Image)
	}
	dig, err := ex.imageStore.Add(location)
	if err != nil {
		return err
	}
	return checkDigest(base, dig)
}

// createArchiveFolder set up a temporary directory for storing all data needed to archive a bundle.
// It sanitizes the name and make sure only the current user has full permission to it.
// If the name contains a path separator, all path separators will be replaced with "-".
func (ex *exporter) createArchiveFolder(name string) (string, error) {
	cleanedPath := strings.ReplaceAll(afero.UnicodeSanitize(name), string(os.PathSeparator), "-")
	archiveDir, err := ex.fs.TempDir("", cleanedPath)
	if err != nil {
		return "", fmt.Errorf("can not create a temporary archive folder: %w", err)
	}

	err = ex.fs.Chmod(archiveDir, pkg.FileModeDirectory)
	if err != nil {
		return "", fmt.Errorf("can not change permission for the temporary archive folder: %w", err)
	}
	return archiveDir, nil
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
