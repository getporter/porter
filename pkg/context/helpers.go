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

	"github.com/deislabs/porter/pkg/test"
	"github.com/spf13/afero"
)

type TestContext struct {
	*Context

	input  *bytes.Buffer
	output *bytes.Buffer
	T      *testing.T
}

// NewTestContext initializes a configuration suitable for testing, with the output buffered, and an in-memory file system.
func NewTestContext(t *testing.T) *TestContext {
	// Provide a way for tests to provide and capture stdin and stdout
	input := &bytes.Buffer{}
	output := &bytes.Buffer{}

	// Copy output to the test log simultaneously, use go test -v to see the output
	aggOutput := io.MultiWriter(output, test.Logger{T: t})

	c := &TestContext{
		Context: &Context{
			Debug:      true,
			FileSystem: &afero.Afero{Fs: afero.NewMemMapFs()},
			In:         input,
			Out:        aggOutput,
			Err:        aggOutput,
			NewCommand: NewTestCommand(),
		},
		output: output,
		T:      t,
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

// TODO: Replace these functions with a union file system for test data
func (c *TestContext) AddTestFile(src, dest string) []byte {
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

func (c *TestContext) GetOutput() string {
	return string(c.output.Bytes())
}
