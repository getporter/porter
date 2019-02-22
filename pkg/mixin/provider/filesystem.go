package mixinprovider

import (
	"path/filepath"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/pkg/errors"
)

func NewFileSystem(config *config.Config) *FileSystem {
	return &FileSystem{
		config: config,
	}
}

type FileSystem struct {
	config *config.Config
}

func (p *FileSystem) GetMixins() ([]mixin.Metadata, error) {
	mixinsDir, err := p.config.GetMixinsDir()
	if err != nil {
		return nil, err
	}

	files, err := p.config.FileSystem.ReadDir(mixinsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list the contents of the mixins directory %q", mixinsDir)
	}

	mixins := make([]mixin.Metadata, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		mixinDir := filepath.Join(mixinsDir, file.Name())
		mixins = append(mixins, mixin.Metadata{
			Name:       file.Name(),
			ClientPath: filepath.Join(mixinDir, file.Name()),
			Dir:        mixinDir,
		})
	}

	return mixins, nil
}
