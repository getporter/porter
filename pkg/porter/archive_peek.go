package porter

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// archiveMetadataFiles are the small, fixed-content files written first in
// an archive built by exporter.CustomTar: bundle.json, relocation-mapping.json,
// and the OCI layout's oci-layout + index.json (see issue #2197). Together
// they're enough to resolve every bundle image's digest (via index.json)
// without reading through the potentially many-gigabyte
// artifacts/layout/blobs/ tree.
var archiveMetadataFiles = []string{
	"bundle.json",
	"relocation-mapping.json",
	"artifacts/layout/oci-layout",
	"artifacts/layout/index.json",
}

// maxPeekEntries bounds how far into the tar stream peekArchiveMetadata will
// scan looking for archiveMetadataFiles before giving up. An archive built
// by this version of Porter carries all four within its first ~10 entries;
// older Porter archives, or ones from other tools, may not have them this
// early (or at all) — 64 is generous headroom at negligible cost either way.
const maxPeekEntries = 64

// peekArchiveMetadata reads just enough of source, a gzip-compressed tar
// archive, to write archiveMetadataFiles into dest, without extracting
// artifacts/layout/blobs/. It reports found=false (not an error) if the
// files aren't all located within maxPeekEntries entries — the caller should
// fall back to a full extraction in that case, which keeps this safe for any
// archive that doesn't have this ordering.
func peekArchiveMetadata(source, dest string) (found bool, err error) {
	f, err := os.Open(source)
	if err != nil {
		return false, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return false, err
	}
	defer gz.Close()

	remaining := make(map[string]bool, len(archiveMetadataFiles))
	for _, name := range archiveMetadataFiles {
		remaining[name] = true
	}

	tr := tar.NewReader(gz)
	for i := 0; i < maxPeekEntries && len(remaining) > 0; i++ {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return false, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		name := strings.TrimPrefix(hdr.Name, "./")
		if !remaining[name] {
			continue
		}

		path := filepath.Join(dest, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return false, err
		}
		if err := writeTarEntry(path, tr); err != nil {
			return false, err
		}

		delete(remaining, name)
	}

	return len(remaining) == 0, nil
}

func writeTarEntry(path string, r io.Reader) error {
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, r)
	return err
}
