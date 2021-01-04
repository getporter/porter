package context

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/carolynvs/aferox"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

type TestContext struct {
	*Context

	cleanupDirs []string
	capturedErr *bytes.Buffer
	capturedOut *bytes.Buffer
	T           *testing.T
}

// NewTestContext initializes a configuration suitable for testing, with the
// output buffered, and an in-memory file system, using the specified
// environment variables.
func NewTestContext(t *testing.T) *TestContext {
	// Provide a way for tests to provide and capture stdin and stdout
	// Copy output to the test log simultaneously, use go test -v to see the output
	err := &bytes.Buffer{}
	aggErr := io.MultiWriter(err, test.Logger{T: t})
	out := &bytes.Buffer{}
	aggOut := io.MultiWriter(out, test.Logger{T: t})

	c := &TestContext{
		Context: &Context{
			Debug:      true,
			environ:    getEnviron(),
			FileSystem: aferox.NewAferox("/", afero.NewMemMapFs()),
			In:         &bytes.Buffer{},
			Out:        aggOut,
			Err:        aggErr,
			PlugInDebugContext: &PluginDebugContext{
				DebuggerPort:           "2735",
				RunPlugInInDebugger:    "",
				PlugInWorkingDirectory: "",
			},
		},
		capturedOut: out,
		capturedErr: err,
		T:           t,
	}

	c.NewCommand = NewTestCommand(c.Context)

	return c
}

func NewTestCommand(c *Context) CommandBuilder {
	return func(command string, args ...string) *exec.Cmd {
		testArgs := append([]string{command}, args...)
		cmd := exec.Command(os.Args[0], testArgs...)
		cmd.Dir = c.Getwd()
		cmd.Env = c.Environ()
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=true", test.MockedCommandEnv))

		return cmd
	}
}

func (c *TestContext) GetTestDefinitionDirectory() string {
	for i := 0; true; i++ {
		_, filename, _, ok := runtime.Caller(i)
		if !ok {
			c.T.Fatal("could not determine calling test directory")
		}
		filename = strings.ToLower(filename)
		if strings.HasSuffix(filename, "_test.go") {
			return filepath.Dir(filename)
		}
	}
	return ""
}

// UseFilesystem has porter's context use the OS filesystem instead of an in-memory filesystem
// Returns the test directory, and the temp porter home directory.
func (c *TestContext) UseFilesystem() (testDir string, homeDir string) {
	homeDir, err := ioutil.TempDir("", "porter-test")
	require.NoError(c.T, err)
	c.cleanupDirs = append(c.cleanupDirs, homeDir)

	testDir = c.GetTestDefinitionDirectory()
	c.FileSystem = aferox.NewAferox(testDir, afero.NewOsFs())
	c.defaultNewCommand()

	return testDir, homeDir
}

func (c *TestContext) AddCleanupDir(dir string) {
	c.cleanupDirs = append(c.cleanupDirs, dir)
}

func (c *TestContext) Cleanup() {
	for _, dir := range c.cleanupDirs {
		c.FileSystem.RemoveAll(dir)
	}
}

func (c *TestContext) AddTestFile(src, dest string) []byte {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		c.T.Fatal(errors.Wrapf(err, "error reading file %s from host filesystem", src))
	}

	err = c.FileSystem.WriteFile(dest, data, os.ModePerm)
	if err != nil {
		c.T.Fatal(errors.Wrapf(err, "error writing file %s to test filesystem", dest))
	}

	return data
}

func (c *TestContext) AddTestFileContents(file []byte, dest string) error {
	return c.FileSystem.WriteFile(dest, file, os.ModePerm)
}

func (c *TestContext) AddTestDirectory(srcDir, destDir string) {
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root src directory
		if path == srcDir {
			return nil
		}

		// Translate the path from the src to the final destination
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return c.FileSystem.MkdirAll(dest, os.ModePerm)
		}

		c.AddTestFile(path, dest)
		return nil
	})
	if err != nil {
		c.T.Fatal(err)
	}
}

func (c *TestContext) AddTestDriver(src, name string) string {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		c.T.Fatal(err)
	}

	dirname, err := c.FileSystem.TempDir("", "porter")
	if err != nil {
		c.T.Fatal(err)
	}

	// filename in accordance with cnab-go's command driver expectations
	filename := fmt.Sprintf("%s/cnab-%s", dirname, name)

	newfile, err := c.FileSystem.Create(filename)
	if err != nil {
		c.T.Fatal(err)
	}

	if len(data) > 0 {
		_, err := newfile.Write(data)
		if err != nil {
			c.T.Fatal(err)
		}
	}

	err = c.FileSystem.Chmod(newfile.Name(), os.ModePerm)
	if err != nil {
		c.T.Fatal(err)
	}
	err = newfile.Close()
	if err != nil {
		c.T.Fatal(err)
	}

	path := c.Getenv("PATH")
	pathlist := []string{dirname, path}
	newpath := strings.Join(pathlist, string(os.PathListSeparator))
	c.Setenv("PATH", newpath)

	return dirname
}

// GetOutput returns all text printed to stdout.
func (c *TestContext) GetOutput() string {
	return string(c.capturedOut.Bytes())
}

// GetError returns all text printed to stderr.
func (c *TestContext) GetError() string {
	return string(c.capturedErr.Bytes())
}

func (c *TestContext) ClearOutputs() {
	c.capturedOut.Truncate(0)
	c.capturedErr.Truncate(0)
}

func (c *TestContext) FindBinDir() string {
	var binDir string
	d := c.GetTestDefinitionDirectory()
	for {
		binDir = c.getBinDir(d)
		if binDir != "" {
			return binDir
		}

		d = filepath.Dir(d)
		if d == "." || d == "" {
			c.T.Fatal("could not find the bin directory")
		}
	}
}

func (c *TestContext) getBinDir(dir string) string {
	children, err := ioutil.ReadDir(dir)
	if err != nil {
		c.T.Fatal(err)
	}
	for _, child := range children {
		if child.IsDir() && child.Name() == "bin" {
			return filepath.Join(dir, child.Name())
		}
	}
	return ""
}
