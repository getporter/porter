package context

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

type TestContext struct {
	*Context

	output *bytes.Buffer
	T      *testing.T
}

// NewTestContext initializes a configuration suitable for testing, with the output buffered, and an in-memory file system.
func NewTestContext(t *testing.T) *TestContext {
	output := &bytes.Buffer{}
	c := &TestContext{
		Context: &Context{
			FileSystem: &afero.Afero{Fs: afero.NewMemMapFs()},
			Out:        output,
			Err:        output,
		},
		output: output,
		T:      t,
	}

	return c
}

func (c *TestContext) AddFile(src, dest string) []byte {
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

func (c *TestContext) AddDirectory(srcDir, destDir string) {
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

		c.AddFile(path, dest)
		return nil
	})
	if err != nil {
		c.T.Fatal(err)
	}
}

func (c *TestContext) GetOutput() string {
	return string(c.output.Bytes())
}
