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
// exporter.export produces: bundle.json + relocation-mapping.json alongside
// an artifacts/ dir holding a large blob (standing in for real OCI layers).
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

	blobDir := filepath.Join(dir, "artifacts", "layout", "blobs", "sha256")
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
