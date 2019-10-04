package packager

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/loader"
	"github.com/docker/docker/pkg/archive"
)

var (
	// ErrNoArtifactsDirectory indicates a missing artifacts/ directory
	ErrNoArtifactsDirectory = errors.New("No artifacts/ directory found")
)

// Importer is responsible for importing a file
type Importer struct {
	Source      string
	Destination string
	Loader      loader.BundleLoader
	Verbose     bool
}

// NewImporter creates a new secure *Importer
//
// source is the filesystem path to the archive.
// destination is the directory to unpack the contents.
// load is a loader.BundleLoader preconfigured for loading bundles.
func NewImporter(source, destination string, load loader.BundleLoader, verbose bool) (*Importer, error) {
	return &Importer{
		Source:      source,
		Destination: destination,
		Loader:      load,
		Verbose:     verbose,
	}, nil
}

// Import decompresses a bundle from Source (location of the compressed bundle) and properly places artifacts in the correct location(s)
func (im *Importer) Import() error {
	_, _, err := im.Unzip()

	// TODO: https://github.com/deislabs/duffle/issues/758

	return err
}

// Unzip decompresses a bundle from Source (location of the compressed bundle) and returns the path of the bundle and the bundle itself.
func (im *Importer) Unzip() (string, *bundle.Bundle, error) {
	baseDir := strings.TrimSuffix(filepath.Base(im.Source), ".tgz")
	dest := filepath.Join(im.Destination, baseDir)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return "", nil, err
	}

	reader, err := os.Open(im.Source)
	if err != nil {
		return "", nil, err
	}
	defer reader.Close()

	tarOptions := &archive.TarOptions{
		Compression:      archive.Gzip,
		IncludeFiles:     []string{"."},
		IncludeSourceDir: true,
		// Issue #416
		NoLchown: true,
	}
	if err := archive.Untar(reader, dest, tarOptions); err != nil {
		return "", nil, fmt.Errorf("untar failed: %s", err)
	}

	// We try to load a bundle.cnab file first, and fall back to a bundle.json
	ext := "cnab"
	if _, err := os.Stat(filepath.Join(dest, "bundle.cnab")); os.IsNotExist(err) {
		ext = "json"
	}

	bun, err := im.Loader.Load(filepath.Join(dest, "bundle."+ext))
	if err != nil {
		removeErr := os.RemoveAll(dest)
		if removeErr != nil {
			return "", nil, fmt.Errorf("failed to load and validate bundle.%s on import %s and failed to remove invalid bundle from filesystem %s", ext, err, removeErr)
		}
		return "", nil, fmt.Errorf("failed to load and validate bundle.%s: %s", ext, err)
	}
	return dest, bun, nil
}
