package porter

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	cnabtooci "get.porter.sh/porter/pkg/cnab/cnab-to-oci"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/carolynvs/aferox"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/imagestore"
	"github.com/cnabio/cnab-go/imagestore/construction"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/spf13/afero"
)

// ArchiveOptions defines the valid options for performing an archive operation
type ArchiveOptions struct {
	BundleReferenceOptions
	ArchiveFile string
}

// Validate performs validation on the publish options
func (o *ArchiveOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	if len(args) < 1 || args[0] == "" {
		return errors.New("destination file is required")
	}
	if len(args) > 1 {
		return fmt.Errorf("only one positional argument may be specified, the archive file name, but multiple were received: %s", args)
	}
	o.ArchiveFile = args[0]

	if o.Reference == "" {
		return errors.New("must provide a value for --reference of the form REGISTRY/bundle:tag")
	}
	return o.BundleReferenceOptions.Validate(ctx, args, p)
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

	bundleRef, err := p.resolveBundleReference(ctx, &opts.BundleReferenceOptions)
	if err != nil {
		return log.Error(err)
	}

	// This allows you to export thin or thick bundles, we only support generating "thick" archives
	ctor, err := construction.NewConstructor(false)
	if err != nil {
		return log.Error(err)
	}

	dest, err := p.Config.FileSystem.OpenFile(opts.ArchiveFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, pkg.FileModeWritable)
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
		insecureRegistry:      opts.InsecureRegistry,
	}
	if err := exp.export(ctx); err != nil {
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
	insecureRegistry      bool
}

func (ex *exporter) export(ctx context.Context) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

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
		return fmt.Errorf("unable to write bundle.json in archive: %w", err)
	}

	reloData, err := json.Marshal(ex.relocationMap)
	if err != nil {
		return err
	}
	err = ex.fs.WriteFile(filepath.Join(archiveDir, "relocation-mapping.json"), reloData, pkg.FileModeWritable)
	if err != nil {
		return fmt.Errorf("unable to write relocation-mapping.json in archive: %w", err)
	}

	var transport *http.Transport
	if ex.insecureRegistry {
		transport = cnabtooci.GetInsecureRegistryTransport()
	} else {
		transport = http.DefaultTransport.(*http.Transport)
	}

	ex.imageStore, err = ex.imageStoreConstructor(
		imagestore.WithArchiveDir(archiveDir),
		imagestore.WithLogs(ex.logs),
		imagestore.WithTransport(transport))
	if err != nil {
		return fmt.Errorf("error creating artifacts: %s", err)
	}

	if err := ex.prepareArtifacts(ex.bundle); err != nil {
		return fmt.Errorf("error preparing bundle artifact: %s", err)
	}

	rc, err := ex.CustomTar(ctx, archiveDir)
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.Copy(ex.destination, rc)
	return err
}

func (ex *exporter) createTarHeader(ctx context.Context, path string, file string, fileInfo os.FileInfo) (*tar.Header, error) {
	log := tracing.LoggerFromContext(ctx)

	header := &tar.Header{
		ModTime:    time.Unix(0, 0),
		AccessTime: time.Unix(0, 0),
		ChangeTime: time.Unix(0, 0),
		Uid:        0,
		Gid:        0,
	}

	switch {
	case fileInfo.Mode().IsDir():
		header.Typeflag = tar.TypeDir
		header.Mode = 0755
	case fileInfo.Mode().IsRegular():
		header.Typeflag = tar.TypeReg
		header.Mode = 0644
		header.Size = fileInfo.Size()
	default:
		log.Debugf("Skipping %s. Not a file/dir", file)
		return nil, nil
	}

	// ensure header has relative file path prepended with '.'
	relativeFilePathName := file

	if filepath.IsAbs(path) {
		relativePath, err := filepath.Rel(path, file)

		if err != nil {
			return nil, err
		}

		if relativePath != "." {
			relativeFilePathName = fmt.Sprintf(".%s%s", string(filepath.Separator), relativePath)
		} else {
			relativeFilePathName = relativePath
		}
	}

	log.Debugf("relativeFilePathName: %s\n", relativeFilePathName)

	header.Name = filepath.ToSlash(relativeFilePathName)

	// directories must be suffixed with '/'
	if fileInfo.Mode().IsDir() && !strings.HasSuffix(header.Name, "/") {
		header.Name += "/"
	}

	log.Debugf("header.Name: %s\n", header.Name)

	return header, nil
}

func (ex *exporter) CustomTar(ctx context.Context, srcPath string) (io.ReadCloser, error) {
	pipeReader, pipeWriter := io.Pipe()

	gzipWriter := gzip.NewWriter(pipeWriter)
	tarWriter := tar.NewWriter(gzipWriter)

	path := filepath.Clean(filepath.Join(srcPath, "."))

	go func() {
		ctx, log := tracing.StartSpan(ctx)

		defer func() {
			if err := tarWriter.Close(); err != nil {
				log.Warnf("Can't close tar writer: %s", err)
			}
			if err := gzipWriter.Close(); err != nil {
				log.Warnf("Can't close gzip writer: %s\n", err)
			}
			if err := pipeWriter.Close(); err != nil {
				log.Warnf("Can't close pipe writer: %s\n", err)
			}
			log.EndSpan()
		}()

		walker := func(file string, finfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			log.Debugf("Packaging directory %s\n", file)

			hdr, err := ex.createTarHeader(ctx, path, file, finfo)
			if err != nil {
				return err
			}

			// if header is nil then it's not a regular file nor directory
			if hdr == nil {
				return nil
			}

			if err := tarWriter.WriteHeader(hdr); err != nil {
				return err
			}

			// if path is a dir, nothing more to do
			if finfo.Mode().IsDir() {
				return nil
			}

			// add file to tar
			sourceFile, err := os.Open(file)
			if err != nil {
				return err
			}

			defer sourceFile.Close()
			_, err = io.Copy(tarWriter, sourceFile)
			if err != nil {
				return err
			}

			return nil
		}

		// build tar
		filepath.Walk(path, walker)
	}()

	return pipeReader, nil
}

// prepareArtifacts pulls all images, verifies their digests and
// saves them to a directory called artifacts/ in the bundle directory
func (ex *exporter) prepareArtifacts(bun cnab.ExtendedBundle) error {
	var imageKeys []string
	for imageKey := range bun.Images {
		imageKeys = append(imageKeys, imageKey)
	}
	sort.Strings(imageKeys)
	for _, k := range imageKeys {
		if err := ex.addImage(bun.Images[k].BaseImage); err != nil {
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
