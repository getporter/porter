package context

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/test"
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

// NewTestContext initializes a configuration suitable for testing, with the output buffered, and an in-memory file system.
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
			FileSystem: &afero.Afero{Fs: afero.NewMemMapFs()},
			In:         &bytes.Buffer{},
			Out:        aggOut,
			Err:        aggErr,
			NewCommand: NewTestCommand(),
		},
		capturedOut: out,
		capturedErr: err,
		T:           t,
	}

	return c
}

func NewTestCommand() CommandBuilder {
	return func(command string, args ...string) *exec.Cmd {
		testArgs := append([]string{command}, args...)
		cmd := exec.Command(os.Args[0], testArgs...)

		cmd.Env = []string{
			fmt.Sprintf("%s=true", test.MockedCommandEnv),
			fmt.Sprintf("%s=%s", test.ExpectedCommandEnv, os.Getenv(test.ExpectedCommandEnv)),
		}

		return cmd
	}
}

// UseFilesystem has porter's context use the OS filesystem instead of an in-memory filesystem
// Returns the temp porter home directory created for the test
func (c *TestContext) UseFilesystem() string {
	c.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}

	testDir, err := ioutil.TempDir("/tmp", "porter")
	require.NoError(c.T, err)
	c.cleanupDirs = append(c.cleanupDirs, testDir)

	return testDir
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
	c.T.Helper()

	data, err := ioutil.ReadFile(src)
	if err != nil {
		c.T.Fatal(err)
	}

	err = c.FileSystem.WriteFile(dest, data, os.ModePerm)
	if err != nil {
		c.T.Fatal(err)
	}

	return data
}

func (c *TestContext) AddTestFileContents(file []byte, dest string) error {
	return c.FileSystem.WriteFile(dest, file, os.ModePerm)
}

func (c *TestContext) AddTestDirectory(srcDir, destDir string) {
	c.T.Helper()

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root src directory
		if path == srcDir {
			return nil
		}

		// Translate the path from the src to the final destination
		dest := filepath.Join(destDir, strings.TrimPrefix(path, srcDir))

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
	c.T.Helper()

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
		newfile.Write(data)
	}

	c.FileSystem.Chmod(newfile.Name(), os.ModePerm)
	newfile.Close()
	path := os.Getenv("PATH")
	pathlist := []string{dirname, path}
	newpath := strings.Join(pathlist, string(os.PathListSeparator))
	os.Setenv("PATH", newpath)

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
	d, err := os.Getwd()
	if err != nil {
		c.T.Fatal(err)
	}
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
