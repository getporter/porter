package porter

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/rand"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkTimeToBundleJSON measures how much of the archive stream must be
// read (decompressed and scanned) before bundle.json's tar header is found.
// This simulates a client that streams a .tgz sequentially without
// extracting it to disk first — see issue #2197, where bundle.json used to
// be the last entry in the tar, after all of artifacts/.
func BenchmarkTimeToBundleJSON(b *testing.B) {
	srcPath := buildBenchArchiveDir(b, 8*1024*1024)
	ex := &exporter{}

	for b.Loop() {
		rc, err := ex.CustomTar(context.Background(), srcPath, gzip.DefaultCompression)
		if err != nil {
			b.Fatal(err)
		}
		if err := scanUntilBundleJSON(rc); err != nil {
			b.Fatal(err)
		}
	}
}

// buildBenchArchiveDir lays out an archive staging dir the same shape as
// exporter.export produces: bundle.json + relocation-mapping.json, and an
// artifacts/layout/ dir with oci-layout + index.json alongside a large blob
// under blobs/ (standing in for real OCI layers).
func buildBenchArchiveDir(b *testing.B, blobSize int64) string {
	b.Helper()
	dir := b.TempDir()

	writeFile := func(name string, data []byte) {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			b.Fatal(err)
		}
	}
	writeFile("bundle.json", []byte(`{"schemaVersion":"1.0.0","name":"bench-bundle","version":"0.1.0"}`))
	writeFile("relocation-mapping.json", []byte(`{}`))

	layoutDir := filepath.Join(dir, "artifacts", "layout")
	if err := os.MkdirAll(layoutDir, 0755); err != nil {
		b.Fatal(err)
	}
	writeFile(filepath.Join("artifacts", "layout", "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`))
	writeFile(filepath.Join("artifacts", "layout", "index.json"), []byte(`{"schemaVersion":2,"manifests":[]}`))

	blobDir := filepath.Join(layoutDir, "blobs", "sha256")
	if err := os.MkdirAll(blobDir, 0755); err != nil {
		b.Fatal(err)
	}
	blob, err := os.Create(filepath.Join(blobDir, "0000000000000000000000000000000000000000000000000000000000000000"))
	if err != nil {
		b.Fatal(err)
	}
	defer blob.Close()
	if _, err := io.CopyN(blob, rand.Reader, blobSize); err != nil {
		b.Fatal(err)
	}

	return dir
}

// buildBenchArchiveFile tars srcDir (via exporter.CustomTar) into a real
// .tgz file, the shape peekArchiveMetadata and extractBundle both expect.
func buildBenchArchiveFile(b *testing.B, srcDir string) string {
	b.Helper()
	ex := &exporter{}
	rc, err := ex.CustomTar(context.Background(), srcDir, gzip.DefaultCompression)
	if err != nil {
		b.Fatal(err)
	}
	defer rc.Close()

	path := filepath.Join(b.TempDir(), "bundle.tgz")
	out, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	defer out.Close()
	if _, err := out.ReadFrom(rc); err != nil {
		b.Fatal(err)
	}
	return path
}

// BenchmarkExtractBundleFull measures today's cost of getting a bundle's
// metadata out of an archive via extractBundle: cnab-go's Importer fully
// extracts every entry, including artifacts/layout/blobs/, regardless of
// whether any of it is actually needed.
func BenchmarkExtractBundleFull(b *testing.B) {
	srcDir := buildBenchArchiveDir(b, 8*1024*1024)
	source := buildBenchArchiveFile(b, srcDir)
	p := &Porter{}

	for b.Loop() {
		tmpDir := b.TempDir()
		if _, err := p.extractBundle(context.Background(), tmpDir, source); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkPeekArchiveMetadata measures the fast path's cost: reading just
// the small metadata files needed to check whether every image is already
// published (see allImagesAlreadyPublished), without extracting
// artifacts/layout/blobs/ at all.
func BenchmarkPeekArchiveMetadata(b *testing.B) {
	srcDir := buildBenchArchiveDir(b, 8*1024*1024)
	source := buildBenchArchiveFile(b, srcDir)

	for b.Loop() {
		dest := b.TempDir()
		found, err := peekArchiveMetadata(source, dest)
		if err != nil {
			b.Fatal(err)
		}
		if !found {
			b.Fatal("expected to find all metadata files within maxPeekEntries")
		}
	}
}

// scanUntilBundleJSON reads the tar.gz stream until bundle.json's header
// appears, then drains the remainder in the background so CustomTar's
// producer goroutine isn't left blocked writing to a full pipe.
func scanUntilBundleJSON(r io.ReadCloser) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	tr := tar.NewReader(gz)

	for {
		hdr, err := tr.Next()
		if err != nil {
			return err
		}
		if hdr.Name == "bundle.json" || hdr.Name == "./bundle.json" {
			go func() {
				_, _ = io.Copy(io.Discard, gz)
				_ = gz.Close()
				_ = r.Close()
			}()
			return nil
		}
	}
}
