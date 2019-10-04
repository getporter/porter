package packager

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/loader"
	"github.com/deislabs/cnab-go/imagestore"
	"github.com/docker/docker/pkg/archive"
)

type Exporter struct {
	source                string
	destination           string
	imageStoreConstructor imagestore.Constructor
	imageStore            imagestore.Store
	logs                  string
	loader                loader.BundleLoader
}

// NewExporter returns an *Exporter given information about where a bundle
//  lives, where the compressed bundle should be exported to. It also
//  sets up a docker client to work with images.
func NewExporter(source, dest, logsDir string, l loader.BundleLoader, c imagestore.Constructor) (*Exporter, error) {
	logs := filepath.Join(logsDir, "export-"+time.Now().Format("20060102150405"))

	return &Exporter{
		source:                source,
		destination:           dest,
		imageStoreConstructor: c,
		logs:                  logs,
		loader:                l,
	}, nil
}

// Export prepares an artifacts directory containing all of the necessary
//  images, packages the bundle along with the artifacts in a gzipped tar
//  file, and saves that file to the file path specified as destination.
//  If the any part of the destination path doesn't, it will be created.
//  exist
func (ex *Exporter) Export() error {
	//prepare log file for this export
	logsf, err := os.Create(ex.logs)
	if err != nil {
		return err
	}
	defer logsf.Close()

	fi, err := os.Stat(ex.source)
	if os.IsNotExist(err) {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("Bundle manifest %s is a directory, should be a file", ex.source)
	}

	bun, err := ex.loader.Load(ex.source)
	if err != nil {
		return fmt.Errorf("Error loading bundle: %s", err)
	}
	name := bun.Name + "-" + bun.Version
	archiveDir, err := ioutil.TempDir("", name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}
	defer os.RemoveAll(archiveDir)

	from, err := os.Open(ex.source)
	if err != nil {
		return err
	}
	defer from.Close()

	bundlefile := "bundle.json"
	to, err := os.OpenFile(filepath.Join(archiveDir, bundlefile), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}

	ex.imageStore, err = ex.imageStoreConstructor(imagestore.WithArchiveDir(archiveDir), imagestore.WithLogs(logsf))
	if err != nil {
		return fmt.Errorf("Error creating artifacts: %s", err)
	}

	if err := ex.prepareArtifacts(bun); err != nil {
		return fmt.Errorf("Error preparing artifacts: %s", err)
	}

	dest := name + ".tgz"
	if ex.destination != "" {
		dest = ex.destination
	}

	writer, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("Error creating archive file: %s", err)
	}

	defer writer.Close()

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

	_, err = io.Copy(writer, rc)
	return err
}

// prepareArtifacts pulls all images, verifies their digests and
// saves them to a directory called artifacts/ in the bundle directory
func (ex *Exporter) prepareArtifacts(bun *bundle.Bundle) error {
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
func (ex *Exporter) addImage(image bundle.BaseImage) error {
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

func (ex *Exporter) Logs() string {
	return ex.logs
}
