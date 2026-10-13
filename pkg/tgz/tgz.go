// Package tgz packs and unpacks directories as gzipped tar archives.
package tgz

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const dirMode = 0o755

// PackDir writes the contents of srcDir to a tar.gz archive. Entry names are
// relative to srcDir. Only regular files and directories are supported,
// symlinks and other file types result in an error.
func PackDir(fsys afero.Fs, srcDir string) ([]byte, error) {
	info, err := fsys.Stat(srcDir)
	if err != nil {
		return nil, fmt.Errorf("could not pack directory %s: %w", srcDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("could not pack directory %s: it is not a directory", srcDir)
	}

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	err = afero.Walk(fsys, srcDir, func(p string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, p)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		// Lstat-based filesystems report symlinks via the mode
		if !info.Mode().IsRegular() && !info.IsDir() {
			return fmt.Errorf("unsupported file type %s for %s: only regular files and directories can be packed", info.Mode().Type(), p)
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("could not create tar header for %s: %w", p, err)
		}
		header.Name = filepath.ToSlash(rel)
		if info.IsDir() {
			header.Name += "/"
		}
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("could not write tar header for %s: %w", p, err)
		}
		if info.IsDir() {
			return nil
		}

		f, err := fsys.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(tw, f); err != nil {
			return fmt.Errorf("could not write %s to archive: %w", p, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not pack directory %s: %w", srcDir, err)
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnpackDir extracts a tar.gz archive created by PackDir into destDir,
// creating it when necessary. Entries that would be written outside of
// destDir are rejected.
func UnpackDir(fsys afero.Fs, archive io.Reader, destDir string) error {
	gr, err := gzip.NewReader(archive)
	if err != nil {
		return fmt.Errorf("could not read gzip archive: %w", err)
	}
	defer gr.Close()

	if err := fsys.MkdirAll(destDir, dirMode); err != nil {
		return fmt.Errorf("could not create directory %s: %w", destDir, err)
	}

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("could not read tar archive: %w", err)
		}

		name := path.Clean(header.Name)
		if path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") {
			return fmt.Errorf("invalid path %q in archive: it would be extracted outside of %s", header.Name, destDir)
		}
		if name == "." {
			continue
		}
		dest := filepath.Join(destDir, filepath.FromSlash(name))

		switch header.Typeflag {
		case tar.TypeDir:
			if err := fsys.MkdirAll(dest, dirMode); err != nil {
				return fmt.Errorf("could not create directory %s: %w", dest, err)
			}
		case tar.TypeReg:
			if err := fsys.MkdirAll(filepath.Dir(dest), dirMode); err != nil {
				return fmt.Errorf("could not create directory %s: %w", filepath.Dir(dest), err)
			}
			f, err := fsys.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.FileMode(header.Mode).Perm())
			if err != nil {
				return fmt.Errorf("could not create file %s: %w", dest, err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return fmt.Errorf("could not write file %s: %w", dest, err)
			}
			if err := f.Close(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported entry type %q for %s in archive", header.Typeflag, header.Name)
		}
	}
}

// Summary describes the contents of an archive created by PackDir.
type Summary struct {
	// Files is the number of regular files in the archive.
	Files int

	// Size is the total size in bytes of the regular files, when extracted.
	Size int64
}

// Summarize reads a tar.gz archive and counts the files it contains.
func Summarize(archive io.Reader) (Summary, error) {
	var s Summary
	gr, err := gzip.NewReader(archive)
	if err != nil {
		return s, fmt.Errorf("could not read gzip archive: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return s, nil
		}
		if err != nil {
			return s, fmt.Errorf("could not read tar archive: %w", err)
		}
		if header.Typeflag == tar.TypeReg {
			s.Files++
			s.Size += header.Size
		}
	}
}

// String formats the summary for display, e.g. "3 files, 1.2 KB".
func (s Summary) String() string {
	noun := "files"
	if s.Files == 1 {
		noun = "file"
	}
	return fmt.Sprintf("%d %s, %s", s.Files, noun, humanSize(s.Size))
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}
