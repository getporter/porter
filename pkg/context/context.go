package context

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// MixinOutputsDir represents the directory where mixin output files are written/read
	MixinOutputsDir = "/cnab/app/porter/outputs"
)

type CommandBuilder func(name string, arg ...string) *exec.Cmd

type Context struct {
	Debug      bool
	verbose    bool
	FileSystem *afero.Afero
	In         io.Reader
	Out        io.Writer
	Err        io.Writer
	NewCommand CommandBuilder
}

func (c *Context) SetVerbose(value bool) {
	c.verbose = value
}

func (c *Context) IsVerbose() bool {
	return c.Debug || c.verbose
}

// CensoredWriter is a writer wrapping the provided io.Writer with logic to censor certain values
type CensoredWriter struct {
	writer          io.Writer
	sensitiveValues []string
}

// NewCensoredWriter returns a new CensoredWriter
func NewCensoredWriter(writer io.Writer) *CensoredWriter {
	return &CensoredWriter{writer: writer, sensitiveValues: []string{}}
}

// SetSensitiveValues sets values needing masking for an CensoredWriter
func (cw *CensoredWriter) SetSensitiveValues(vals []string) {
	cw.sensitiveValues = vals
}

// Write implements io.Writer's Write method, performing necessary auditing while doing so
func (cw *CensoredWriter) Write(b []byte) (int, error) {
	auditedBytes := b
	for _, val := range cw.sensitiveValues {
		if strings.TrimSpace(val) != "" {
			auditedBytes = bytes.Replace(auditedBytes, []byte(val), []byte("*******"), -1)
		}
	}

	_, err := cw.writer.Write(auditedBytes)
	return len(b), err
}

func New() *Context {
	// Default to respecting the PORTER_DEBUG env variable, the cli will override if --debug is set otherwise
	_, debug := os.LookupEnv("PORTER_DEBUG")

	return &Context{
		Debug:      debug,
		FileSystem: &afero.Afero{Fs: afero.NewOsFs()},
		In:         os.Stdin,
		Out:        NewCensoredWriter(os.Stdout),
		Err:        NewCensoredWriter(os.Stderr),
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

// WriteMixinOutputToFile writes the provided bytes (representing a mixin output)
// to a file named by the provided filename in Porter's mixin outputs directory
func (c *Context) WriteMixinOutputToFile(filename string, bytes []byte) error {
	exists, err := c.FileSystem.DirExists(MixinOutputsDir)
	if err != nil {
		return err
	}
	if !exists {
		if err := c.FileSystem.MkdirAll(MixinOutputsDir, os.ModePerm); err != nil {
			return errors.Wrap(err, "couldn't make output directory")
		}
	}

	return c.FileSystem.WriteFile(filepath.Join(MixinOutputsDir, filename), bytes, os.ModePerm)
}

// SetSensitiveValues sets the sensitive values needing masking on output/err streams
func (c *Context) SetSensitiveValues(vals []string) {
	if len(vals) > 0 {
		out := NewCensoredWriter(c.Out)
		out.SetSensitiveValues(vals)
		c.Out = out

		err := NewCensoredWriter(c.Err)
		err.SetSensitiveValues(vals)
		c.Err = err
	}
}
