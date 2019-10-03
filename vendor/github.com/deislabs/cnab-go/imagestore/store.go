package imagestore

import (
	"io"
	"io/ioutil"

	"github.com/pivotal/image-relocation/pkg/image"
)

// Store is an abstract image store.
type Store interface {
	// Add copies the image with the given name to the image store.
	Add(img string) (contentDigest string, err error)

	// Push copies the image with the given digest from an image with the given name in the image store to a repository
	// with the given name.
	Push(dig image.Digest, src image.Name, dst image.Name) error
}

// Constructor is a function which creates an images store based on parameters represented as options
type Constructor func(...Option) (Store, error)

// Parameters is used to create image stores.
type Parameters struct {
	ArchiveDir string
	Logs       io.Writer
}

// Options is a function which returns updated parameters.
type Option func(Parameters) Parameters

func Create(options ...Option) Parameters {
	b := Parameters{
		Logs: ioutil.Discard,
	}
	for _, op := range options {
		b = op(b)
	}
	return b
}

// WithArchiveDir return an option to set the archive directory parameter.
func WithArchiveDir(archiveDir string) Option {
	return func(b Parameters) Parameters {
		return Parameters{
			ArchiveDir: archiveDir,
			Logs:       b.Logs,
		}
	}
}

// WithArchiveDir return an option to set the logs parameter.
func WithLogs(logs io.Writer) Option {
	return func(b Parameters) Parameters {
		return Parameters{
			ArchiveDir: b.ArchiveDir,
			Logs:       logs,
		}
	}
}
