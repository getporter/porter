package context

import (
	"bytes"
	"io"
	"os"

	"github.com/spf13/afero"
)

type Context struct {
	Debug      bool
	FileSystem *afero.Afero
	Out        io.Writer
}

func New() Context {
	return Context{
		FileSystem: &afero.Afero{Fs: afero.NewOsFs()},
		Out:        os.Stdout,
	}
}

// NewTestContext initializes a configuration suitable for testing, with the output buffered, and an in-memory file system.
func NewTestContext() (Context, *bytes.Buffer) {
	output := &bytes.Buffer{}
	c := Context{
		FileSystem: &afero.Afero{Fs: afero.NewMemMapFs()},
		Out:        output,
	}

	return c, output
}
