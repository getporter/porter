package context

import (
	"io"
	"os"

	"github.com/spf13/afero"
)

type Context struct {
	Debug      bool
	FileSystem *afero.Afero
	Out        io.Writer
	Err        io.Writer
}

func New() *Context {
	return &Context{
		FileSystem: &afero.Afero{Fs: afero.NewOsFs()},
		Out:        os.Stdout,
		Err:        os.Stderr,
	}
}
