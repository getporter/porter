package tgz

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackUnpackDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fsys, "/src/a.txt", []byte("a"), 0o644))
	require.NoError(t, afero.WriteFile(fsys, "/src/sub/deep/b.sh", []byte("b"), 0o755))
	require.NoError(t, fsys.MkdirAll("/src/empty", 0o755))

	data, err := PackDir(fsys, "/src")
	require.NoError(t, err)

	require.NoError(t, UnpackDir(fsys, bytes.NewReader(data), "/dest"))

	got, err := afero.ReadFile(fsys, "/dest/a.txt")
	require.NoError(t, err)
	assert.Equal(t, "a", string(got))
	got, err = afero.ReadFile(fsys, "/dest/sub/deep/b.sh")
	require.NoError(t, err)
	assert.Equal(t, "b", string(got))
	info, err := fsys.Stat("/dest/sub/deep/b.sh")
	require.NoError(t, err)
	assert.Equal(t, 0o755, int(info.Mode().Perm()))
	ok, err := afero.IsDir(fsys, "/dest/empty")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestUnpackDir_RejectsTraversal(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "../evil", Mode: 0o644, Size: 1, Typeflag: tar.TypeReg}))
	_, err := tw.Write([]byte("x"))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gw.Close())

	fsys := afero.NewMemMapFs()
	err = UnpackDir(fsys, &buf, "/dest")
	require.ErrorContains(t, err, "outside of")
	exists, _ := afero.Exists(fsys, "/evil")
	assert.False(t, exists)
}

func TestSummarize(t *testing.T) {
	fsys := afero.NewMemMapFs()
	require.NoError(t, afero.WriteFile(fsys, "/src/a.txt", []byte("12345"), 0o644))
	require.NoError(t, afero.WriteFile(fsys, "/src/sub/b.txt", make([]byte, 2048), 0o644))
	data, err := PackDir(fsys, "/src")
	require.NoError(t, err)

	s, err := Summarize(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, 2, s.Files)
	assert.Equal(t, "2 files, 2.0 KB", s.String())

	_, err = Summarize(bytes.NewReader([]byte("not an archive")))
	require.Error(t, err)

	assert.Equal(t, "1 file, 5 B", Summary{Files: 1, Size: 5}.String())
}

func TestPackUnpackDir_Empty(t *testing.T) {
	fsys := afero.NewMemMapFs()
	require.NoError(t, fsys.MkdirAll("/src", 0o755))

	data, err := PackDir(fsys, "/src")
	require.NoError(t, err)

	require.NoError(t, UnpackDir(fsys, bytes.NewReader(data), "/dest"))
	ok, err := afero.IsDir(fsys, "/dest")
	require.NoError(t, err)
	assert.True(t, ok, "an empty directory should round trip to an empty directory")

	s, err := Summarize(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, "0 files, 0 B", s.String())
}
