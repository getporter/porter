package context

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/carolynvs/aferox"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// MixinOutputsDir represents the directory where mixin output files are written/read
	MixinOutputsDir = "/cnab/app/porter/outputs"
)

type CommandBuilder func(name string, arg ...string) *exec.Cmd

type Context struct {
	Debug        bool
	DebugPlugins bool
	verbose      bool

	// map of environment variables, all keys are upper case
	environ map[string]string

	FileSystem         aferox.Aferox
	In                 io.Reader
	Out                io.Writer
	Err                io.Writer
	NewCommand         CommandBuilder
	PlugInDebugContext *PluginDebugContext
}

// New creates a new context in the specified directory.
func New() *Context {
	// Ignore any error getting the working directory and report errors
	// when we attempt to access files in the current directory. This
	// allows us to use the current directory as a default, and allow
	// tests to override it.
	pwd, _ := os.Getwd()

	c := &Context{
		environ:    getEnviron(),
		FileSystem: aferox.NewAferox(pwd, afero.NewOsFs()),
		In:         os.Stdin,
		Out:        NewCensoredWriter(os.Stdout),
		Err:        NewCensoredWriter(os.Stderr),
	}
	c.defaultNewCommand()
	c.PlugInDebugContext = NewPluginDebugContext(c)
	return c
}

func (c *Context) defaultNewCommand() {
	c.NewCommand = func(name string, arg ...string) *exec.Cmd {
		return c.Command(name, arg...)
	}
}

// Command creates a new exec.Cmd using the context's current directory and environment variables.
func (c *Context) Command(name string, arg ...string) *exec.Cmd {
	cmd := &exec.Cmd{
		Dir:  c.Getwd(),
		Path: name,
		Args: append([]string{name}, arg...),
		Env:  c.Environ(),
	}
	if filepath.Base(name) == name {
		if lp, ok := c.LookPath(name); ok {
			cmd.Path = lp
		}
	}
	return cmd
}

func getEnviron() map[string]string {
	environ := map[string]string{}
	for _, env := range os.Environ() {
		envParts := strings.SplitN(env, "=", 2)
		key := strings.ToUpper(envParts[0])
		value := ""
		if len(envParts) > 1 {
			value = envParts[1]
		}
		environ[key] = value
	}
	return environ
}

func (c *Context) SetVerbose(value bool) {
	c.verbose = value
}

func (c *Context) IsVerbose() bool {
	return c.Debug || c.verbose
}

// Environ returns a copy of strings representing the environment,
// in the form "key=value".
func (c *Context) Environ() []string {
	e := make([]string, 0, len(c.environ))
	for k, v := range c.environ {
		e = append(e, fmt.Sprintf("%s=%s", k, v))
	}
	return e
}

// ExpandEnv replaces ${var} or $var in the string according to the values
// of the current environment variables. References to undefined
// variables are replaced by the empty string.
func (c *Context) ExpandEnv(s string) string {
	return os.Expand(s, func(key string) string { return c.Getenv(key) })
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will be empty if the variable is not present.
// To distinguish between an empty value and an unset value, use LookupEnv.
func (c *Context) Getenv(key string) string {
	key = strings.ToUpper(key)
	return c.environ[key]
}

// This is a simplified exec.LookPath that checks if command is accessible given
// a PATH environment variable.
func (c *Context) LookPath(file string) (string, bool) {
	return c.FileSystem.LookPath(file, c.Getenv("PATH"), c.Getenv("PATHEXT"))
}

// LookupEnv retrieves the value of the environment variable named
// by the key. If the variable is present in the environment the
// value (which may be empty) is returned and the boolean is true.
// Otherwise the returned value will be empty and the boolean will
// be false.
func (c *Context) LookupEnv(key string) (string, bool) {
	key = strings.ToUpper(key)
	value, ok := c.environ[key]
	return value, ok
}

// Setenv sets the value of the environment variable named by the key.
// It returns an error, if any.
func (c *Context) Setenv(key string, value string) {
	key = strings.ToUpper(key)

	if c.environ == nil {
		c.environ = make(map[string]string, 1)
	}

	c.environ[key] = value
}

// Unsetenv unsets a single environment variable.
func (c *Context) Unsetenv(key string) {
	key = strings.ToUpper(key)

	delete(c.environ, key)
}

// Clearenv deletes all environment variables.
func (c *Context) Clearenv() {
	c.environ = make(map[string]string, 0)
}

// Getwd returns a rooted path name corresponding to the current directory.
func (c *Context) Getwd() string {
	return c.FileSystem.Getwd()
}

// Chdir changes the current working directory to the named directory.
func (c *Context) Chdir(dir string) {
	c.FileSystem.Chdir(dir)
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
