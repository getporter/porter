package porter

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/spf13/afero"
	"go.opentelemetry.io/otel/attribute"
)

// ArchiveOptions defines the valid options for performing an archive operation
type ArchiveOptions struct {
	BundleReferenceOptions
	ArchiveFile         string
	CompressionLevel    string
	compressionLevelInt int
}

var compressionLevelValues = map[string]int{
	"NoCompression":      gzip.NoCompression,
	"BestSpeed":          gzip.BestSpeed,
	"BestCompression":    gzip.BestCompression,
	"DefaultCompression": gzip.DefaultCompression,
	"HuffmanOnly":        gzip.HuffmanOnly,
}

func (o *ArchiveOptions) GetCompressionLevelDefault() string {
	return "DefaultCompression"
}

func (p *ArchiveOptions) GetCompressionLevelAllowedValues() []string {
	levels := make([]string, 0, len(compressionLevelValues))
	for level := range compressionLevelValues {
		levels = append(levels, level)
	}
	sort.Strings(levels)
	return levels
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

	if o.CompressionLevel == "" {
		o.CompressionLevel = o.GetCompressionLevelDefault()
	}
	level, ok := compressionLevelValues[o.CompressionLevel]
	if !ok {
		return fmt.Errorf("invalid compression level: %s", o.CompressionLevel)
	}
	o.compressionLevelInt = level

	return o.BundleReferenceOptions.Validate(ctx, args, p)
}

// Archive is a composite function that generates a CNAB thick bundle. It will pull the bundle image, and
// any referenced images locally (if needed), export them to individual layers, generate a bundle.json and
// then generate a gzipped tar archive containing the bundle.json and the images
func (p *Porter) Archive(ctx context.Context, opts ArchiveOptions) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	dir := filepath.Dir(opts.ArchiveFile)
	if _, err := p.FileSystem.Stat(dir); os.IsNotExist(err) {
		return log.Error(fmt.Errorf("parent directory %q does not exist", filepath.ToSlash(dir)))
	}

	bundleRef, err := opts.GetBundleReference(ctx, p)
	if err != nil {
		return log.Error(err)
	}

	dest, err := p.FileSystem.OpenFile(opts.ArchiveFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, pkg.FileModeWritable)
	if err != nil {
		return log.Error(err)
	}

	exp := &exporter{
		fs:               p.FileSystem,
		out:              p.Out,
		bundle:           bundleRef.Definition,
		relocationMap:    bundleRef.RelocationMap,
		destination:      dest,
		registry:         p.Registry,
		insecureRegistry: opts.InsecureRegistry,
		compressionLevel: opts.compressionLevelInt,
	}
	if err := exp.export(ctx); err != nil {
		return log.Error(err)
	}

	return nil
}

type exporter struct {
	fs               aferox.Aferox
	out              io.Writer
	bundle           cnab.ExtendedBundle
	relocationMap    relocation.ImageRelocationMap
	destination      io.Writer
	registry         cnabtooci.RegistryProvider
	imageStore       imageStore
	insecureRegistry bool
	compressionLevel int
}

// ociRefNameAnnotation is the annotation that records the name of an image in
// an OCI image layout.
const ociRefNameAnnotation = "org.opencontainers.image.ref.name"

// imageStore stores the images referenced by a bundle in the archive.
type imageStore interface {
	// Add copies the image with the given name to the image store.
	Add(ctx context.Context, img string) (contentDigest string, err error)
}

// ociLayoutStore is an image store which stores images as an OCI image layout
// in the artifacts/layout directory of the archive.
type ociLayoutStore struct {
	layoutPath layout.Path
	registry   cnabtooci.RegistryProvider
	regOpts    cnabtooci.RegistryOptions
}

func newOCILayoutStore(archiveDir string, registry cnabtooci.RegistryProvider, regOpts cnabtooci.RegistryOptions) (*ociLayoutStore, error) {
	layoutDir := filepath.Join(archiveDir, "artifacts", "layout")
	if err := os.MkdirAll(layoutDir, pkg.FileModeDirectory); err != nil {
		return nil, err
	}

	layoutPath, err := layout.Write(layoutDir, empty.Index)
	if err != nil {
		return nil, err
	}

	return &ociLayoutStore{layoutPath: layoutPath, registry: registry, regOpts: regOpts}, nil
}

// Add pulls the image (or image index) from its registry and appends it to
// the layout, annotated with its fully-qualified name so that it can be found
// again when the archive is published.
func (s *ociLayoutStore) Add(ctx context.Context, img string) (string, error) {
	ref, err := cnab.ParseOCIReference(img)
	if err != nil {
		return "", err
	}

	desc, err := s.registry.GetImageDescriptor(ctx, ref, s.regOpts)
	if err != nil {
		return "", err
	}

	annotations := layout.WithAnnotations(map[string]string{ociRefNameAnnotation: ref.Named.String()})
	if desc.MediaType.IsIndex() {
		idx, err := desc.ImageIndex()
		if err != nil {
			return "", err
		}
		err = s.layoutPath.AppendIndex(idx, annotations)
		if err != nil {
			return "", err
		}
	} else {
		// assume all other media types are images since some images don't set the media type
		img, err := desc.Image()
		if err != nil {
			return "", err
		}
		err = s.layoutPath.AppendImage(img, annotations)
		if err != nil {
			return "", err
		}
	}

	return desc.Digest.String(), nil
}

func (ex *exporter) export(ctx context.Context) error {
	ctx, log := tracing.StartSpan(ctx)
	defer log.EndSpan()

	name := ex.bundle.Name + "-" + ex.bundle.Version
	archiveDir, err := ex.createArchiveFolder(name)
	if err != nil {
		return fmt.Errorf("can not create archive folder: %w", err)
	}
	defer func() {
		err = errors.Join(err, ex.fs.RemoveAll(archiveDir))
	}()

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

	ex.imageStore, err = newOCILayoutStore(archiveDir, ex.registry, cnabtooci.RegistryOptions{InsecureRegistry: ex.insecureRegistry})
	if err != nil {
		return fmt.Errorf("error creating artifacts: %s", err)
	}

	if err := ex.prepareArtifacts(ctx, ex.bundle); err != nil {
		return fmt.Errorf("error preparing bundle artifact: %s", err)
	}

	rc, err := ex.CustomTar(ctx, archiveDir, ex.compressionLevel)
	if err != nil {
		return fmt.Errorf("error creating archive: %w", err)
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
		log.Debug("Skipping header creation. Not a file/dir", attribute.String("createTarHeader.file", file))
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

	header.Name = filepath.ToSlash(relativeFilePathName)

	// directories must be suffixed with '/'
	if fileInfo.Mode().IsDir() && !strings.HasSuffix(header.Name, "/") {
		header.Name += "/"
	}

	log.Debug("Created tar header", attribute.String("createTarHeader.headerName", header.Name))

	return header, nil
}

func (ex *exporter) CustomTar(ctx context.Context, srcPath string, compressionLevel int) (io.ReadCloser, error) {
	pipeReader, pipeWriter := io.Pipe()

	gzipWriter, err := gzip.NewWriterLevel(pipeWriter, compressionLevel)
	if err != nil {
		return nil, err
	}
	tarWriter := tar.NewWriter(gzipWriter)

	cleanSrcPath := filepath.Clean(srcPath)

	go func() {
		ctx, log := tracing.StartSpanWithName(ctx, "CustomTar.Walk")

		defer func() {
			if err := tarWriter.Close(); err != nil {
				log.Warnf("Can't close tar writer: %s", err)
			}
			if err := gzipWriter.Close(); err != nil {
				log.Warnf("Can't close gzip writer: %s\n", err)
			}
			// Propagate the write error (if any) to the reader instead of
			// closing the pipe cleanly, so callers see a failed read rather
			// than a silently truncated archive. CloseWithError always
			// returns nil.
			_ = pipeWriter.CloseWithError(err)
			log.EndSpan()
		}()

		writePath := func(path string, finfo os.FileInfo) error {
			ctx, log := tracing.StartSpanWithName(ctx, "CustomTar.ProcessPath", attribute.String("customTar.path", path))
			defer log.EndSpan()

			hdr, err := ex.createTarHeader(ctx, srcPath, path, finfo)
			if err != nil {
				return fmt.Errorf("failed to create tar header for path %s: %w", path, err)
			}

			// if header is nil then it's not a regular file nor directory
			if hdr == nil {
				return nil
			}

			if err := tarWriter.WriteHeader(hdr); err != nil {
				return fmt.Errorf("failed to write header for path %s: %w", path, err)
			}

			// if path is a dir, nothing more to do
			if finfo.Mode().IsDir() {
				return nil
			}

			// add file to tar
			sourceFile, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open %s: %w", path, err)
			}

			defer sourceFile.Close()
			_, err = io.Copy(tarWriter, sourceFile)
			if err != nil {
				return fmt.Errorf("failed to copy %s: %w", path, err)
			}

			return nil
		}

		// Tracks paths already written by name below, so the catch-all walk
		// at the end doesn't write them twice.
		written := map[string]bool{}

		walker := func(path string, finfo os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("walk invoked with error: %w", err)
			}
			if written[path] {
				return nil
			}
			return writePath(path, finfo)
		}

		// Write the root dir, then the small metadata files (bundle.json,
		// relocation-mapping.json) before walking artifacts/, so a client
		// streaming the archive can read the bundle manifest without first
		// reading through the (potentially many-gigabyte) artifacts/ tree.
		// See https://github.com/getporter/porter/issues/2197.
		err = func() error {
			writeNamed := func(path string) error {
				finfo, err := os.Stat(path)
				if err != nil {
					return fmt.Errorf("failed to stat %s: %w", path, err)
				}
				written[path] = true
				return writePath(path, finfo)
			}

			if err := writeNamed(cleanSrcPath); err != nil {
				return err
			}

			for _, name := range []string{"bundle.json", "relocation-mapping.json"} {
				if err := writeNamed(filepath.Join(cleanSrcPath, name)); err != nil {
					return err
				}
			}

			// Same reasoning one level deeper: artifacts/layout/{oci-layout,
			// index.json} are small files written by newOCILayoutStore up
			// front, and index.json alone is enough to
			// resolve every image's digest (layout.Path.ImageIndex reads
			// only index.json) — so write them before artifacts/layout/blobs/,
			// the large tree that holds the actual image content.
			artifactsDir := filepath.Join(cleanSrcPath, "artifacts")
			layoutDir := filepath.Join(artifactsDir, "layout")
			for _, dir := range []string{artifactsDir, layoutDir} {
				if _, err := os.Stat(dir); err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return fmt.Errorf("failed to stat %s: %w", dir, err)
				}
				if err := writeNamed(dir); err != nil {
					return err
				}
			}
			for _, name := range []string{"oci-layout", "index.json"} {
				path := filepath.Join(layoutDir, name)
				if _, err := os.Stat(path); err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return fmt.Errorf("failed to stat %s: %w", path, err)
				}
				if err := writeNamed(path); err != nil {
					return err
				}
			}

			// Catch-all: walk everything else under the archive dir (blobs/,
			// and anything not explicitly named above) so nothing is
			// silently dropped from the tar if the staging layout changes.
			return filepath.Walk(cleanSrcPath, walker)
		}()
	}()

	return pipeReader, nil
}

// prepareArtifacts pulls all images, verifies their digests and
// saves them to a directory called artifacts/ in the bundle directory
func (ex *exporter) prepareArtifacts(ctx context.Context, bun cnab.ExtendedBundle) error {
	var imageKeys []string
	for imageKey := range bun.Images {
		imageKeys = append(imageKeys, imageKey)
	}
	sort.Strings(imageKeys)
	for _, k := range imageKeys {
		if err := ex.addImage(ctx, bun.Images[k].BaseImage); err != nil {
			return err
		}
	}

	for _, in := range bun.InvocationImages {
		if err := ex.addImage(ctx, in.BaseImage); err != nil {
			return err
		}
	}

	return nil
}

// addImage pulls an image using relocation map, adds it to the artifacts/ directory, and verifies its digest
func (ex *exporter) addImage(ctx context.Context, base bundle.BaseImage) error {
	if ex.relocationMap == nil {
		return errors.New("relocation map is not provided")
	}
	location, ok := ex.relocationMap[base.Image]
	if !ok {
		return fmt.Errorf("can not locate the referenced image: %s", base.Image)
	}
	dig, err := ex.imageStore.Add(ctx, location)
	if err != nil {
		return err
	}
	return checkDigest(base, dig)
}

// createArchiveFolder set up a temporary directory for storing all data needed to archive a bundle.
// It sanitizes the name and make sure only the current user has full permission to it.
// If the name contains a path separator, all path separators will be replaced with "-".
func (ex *exporter) createArchiveFolder(name string) (string, error) {
	cleanedPath := strings.ReplaceAll(afero.UnicodeSanitize(name), "/", "-")
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
