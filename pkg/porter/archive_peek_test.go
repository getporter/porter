package porter

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPeekArchiveMetadata_Found(t *testing.T) {
	srcDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "bundle.json"), []byte(`{"name":"bench"}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "relocation-mapping.json"), []byte(`{}`), 0644))

	layoutDir := filepath.Join(srcDir, "artifacts", "layout")
	require.NoError(t, os.MkdirAll(layoutDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(layoutDir, "oci-layout"), []byte(`{"imageLayoutVersion":"1.0.0"}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(layoutDir, "index.json"), []byte(`{"schemaVersion":2,"manifests":[]}`), 0644))

	blobsDir := filepath.Join(layoutDir, "blobs", "sha256")
	require.NoError(t, os.MkdirAll(blobsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(blobsDir, "abc123"), []byte("a large blob"), 0644))

	ex := &exporter{}
	rc, err := ex.CustomTar(context.Background(), srcDir, gzip.DefaultCompression)
	require.NoError(t, err)
	defer rc.Close()

	source := filepath.Join(t.TempDir(), "bundle.tgz")
	out, err := os.Create(source)
	require.NoError(t, err)
	_, err = out.ReadFrom(rc)
	require.NoError(t, err)
	require.NoError(t, out.Close())

	dest := t.TempDir()
	found, err := peekArchiveMetadata(source, dest)
	require.NoError(t, err)
	require.True(t, found, "expected to find all metadata files within maxPeekEntries")

	bundleJSON, err := os.ReadFile(filepath.Join(dest, "bundle.json"))
	require.NoError(t, err)
	require.Equal(t, `{"name":"bench"}`, string(bundleJSON))

	_, err = os.ReadFile(filepath.Join(dest, "relocation-mapping.json"))
	require.NoError(t, err)
	_, err = os.ReadFile(filepath.Join(dest, "artifacts", "layout", "oci-layout"))
	require.NoError(t, err)
	_, err = os.ReadFile(filepath.Join(dest, "artifacts", "layout", "index.json"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(dest, "artifacts", "layout", "blobs"))
	require.True(t, os.IsNotExist(err), "expected peekArchiveMetadata to never write artifacts/layout/blobs/")
}

func TestPeekArchiveMetadata_FallsBackWhenMetadataIsLate(t *testing.T) {
	// Simulate an archive that doesn't have the #2197 ordering (an older
	// Porter archive, or one from another tool): pad it with filler entries
	// past maxPeekEntries before the metadata files.
	source := filepath.Join(t.TempDir(), "bundle.tgz")
	out, err := os.Create(source)
	require.NoError(t, err)

	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)

	writeEntry := func(name string, content string) {
		hdr := &tar.Header{Name: name, Mode: 0644, Size: int64(len(content))}
		require.NoError(t, tw.WriteHeader(hdr))
		_, err := tw.Write([]byte(content))
		require.NoError(t, err)
	}

	for i := range maxPeekEntries + 10 {
		writeEntry(fmt.Sprintf("filler-%d", i), "x")
	}
	writeEntry("bundle.json", `{}`)
	writeEntry("relocation-mapping.json", `{}`)
	writeEntry("artifacts/layout/oci-layout", `{}`)
	writeEntry("artifacts/layout/index.json", `{}`)

	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	require.NoError(t, out.Close())

	dest := t.TempDir()
	found, err := peekArchiveMetadata(source, dest)
	require.NoError(t, err)
	require.False(t, found, "expected peekArchiveMetadata to give up after maxPeekEntries")
}

func TestSafeJoin(t *testing.T) {
	dest := filepath.Join(string(filepath.Separator), "dest")

	path, err := safeJoin(dest, "bundle.json")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dest, "bundle.json"), path)

	for _, name := range []string{
		"../escape",
		"../../escape",
		"artifacts/../../escape",
	} {
		_, err := safeJoin(dest, name)
		require.Error(t, err, "expected %q to be rejected as a path traversal attempt", name)
	}
}
