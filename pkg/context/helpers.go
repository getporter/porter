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
)

type TestContext struct {
	*Context

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
		// testArgs := append([]string{command}, args...)
		cmd := exec.Command(command, args...)

		cmd.Env = []string{
			fmt.Sprintf("%s=true", test.MockedCommandEnv),
			fmt.Sprintf("%s=%s", test.ExpectedCommandEnv, os.Getenv(test.ExpectedCommandEnv)),
		}

		return cmd
	}
}

// TODO: Replace these functions with a union file system for test data
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

// GetOutput returns all text printed to stdout.
func (c *TestContext) GetOutput() string {
	return string(c.capturedOut.Bytes())
}

// GetError returns all text printed to stderr.
func (c *TestContext) GetError() string {
	return string(c.capturedErr.Bytes())
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
