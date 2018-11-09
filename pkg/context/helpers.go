package context

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
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
	require.NoError(c.T, err)

	err = c.FileSystem.WriteFile(dest, data, os.ModePerm)
	require.NoError(c.T, err)

	return data
}

func (c *TestContext) GetOutput() string {
	return string(c.output.Bytes())
}
