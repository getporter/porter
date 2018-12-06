package context

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type CommandBuilder func(name string, arg ...string) *exec.Cmd

type Context struct {
	Debug      bool
	FileSystem *afero.Afero
	In         io.Reader
	Out        io.Writer
	Err        io.Writer
	NewCommand CommandBuilder
}

func New() *Context {
	// Default to respecting the PORTER_DEBUG env variable, the cli will override if --debug is set otherwise
	_, debug := os.LookupEnv("PORTER_DEBUG")
	return &Context{
		Debug:      debug,
		FileSystem: &afero.Afero{Fs: afero.NewOsFs()},
		In:         os.Stdin,
		Out:        os.Stdout,
		Err:        os.Stderr,
		NewCommand: exec.Command,
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

// WriteOutput writes the given lines to a file in the
// output directory
func (c *Context) WriteOutput(lines []string) error {
	exists, err := c.FileSystem.DirExists("/cnab/app/porter/outputs")
	if err != nil {
		return err
	}
	if !exists {
		if err := c.FileSystem.MkdirAll("/cnab/app/porter/outputs", os.ModePerm); err != nil {
			return errors.Wrap(err, "couldn't make output directory")
		}
	}
	f, err := c.FileSystem.TempFile("/cnab/app/porter/outputs", "mixin-output")
	if err != nil {
		return errors.Wrap(err, "couldn't open outputs file")
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	defer buf.Flush()
	for _, line := range lines {
		// remove any trailing newline, because we will append one
		line = strings.TrimSuffix(line, "\n")
		line = fmt.Sprintf("%s\n", line)
		_, err := buf.Write([]byte(line))
		if err != nil {
			return err
		}
	}
	return nil
}
