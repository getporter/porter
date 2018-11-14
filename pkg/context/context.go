package context

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/spf13/afero"
)

type Context struct {
	Debug      bool
	FileSystem *afero.Afero
	Out        io.Writer
	Err        io.Writer
}

func New() *Context {
	return &Context{
		FileSystem: &afero.Afero{Fs: afero.NewOsFs()},
		Out:        os.Stdout,
		Err:        os.Stderr,
	}
}

func (c *Context) CopyDirectory(srcDir, destDir string, includeBaseDir bool) error {
	var stripPrefix string
	if includeBaseDir {
		stripPrefix = filepath.Dir(srcDir)
	} else {
		stripPrefix = srcDir
	}

	return c.FileSystem.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.WithStack(err)
		}

		// Translate the path from the src to the final destination
		dest := filepath.Join(destDir, strings.TrimPrefix(path, stripPrefix))
		if dest == "" {
			return nil
		}

		if info.IsDir() {
			return errors.WithStack(c.FileSystem.MkdirAll(dest, info.Mode()))
		}

		return c.CopyFile(path, dest)
	})
}

func (c *Context) CopyFile(src, dest string) error {
	info, err := c.FileSystem.Stat(src)
	if err != nil {
		return errors.WithStack(err)
	}

	data, err := c.FileSystem.ReadFile(src)
	if err != nil {
		return errors.WithStack(err)
	}

	err = c.FileSystem.WriteFile(dest, data, info.Mode())
	return errors.WithStack(err)
}
