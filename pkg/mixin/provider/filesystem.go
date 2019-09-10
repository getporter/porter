package mixinprovider

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/context"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/pkg/errors"
)

func NewFileSystem(config *config.Config) *FileSystem {
	return &FileSystem{
		Config: config,
	}
}

type FileSystem struct {
	*config.Config
}

func (p *FileSystem) List() ([]mixin.Metadata, error) {
	mixinsDir, err := p.GetMixinsDir()
	if err != nil {
		return nil, err
	}

	files, err := p.FileSystem.ReadDir(mixinsDir)
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

func (p *FileSystem) GetSchema(m mixin.Metadata) (string, error) {
	r := mixin.NewRunner(m.Name, m.Dir, false)
	r.Command = "schema"

	// Copy the existing context and tweak to pipe the output differently
	mixinSchema := &bytes.Buffer{}
	var mixinContext context.Context
	mixinContext = *p.Context
	mixinContext.Out = mixinSchema
	if !p.Debug {
		mixinContext.Err = ioutil.Discard
	}
	r.Context = &mixinContext

	err := r.Run()
	if err != nil {
		return "", err
	}

	return mixinSchema.String(), nil
}

// GetVersion is the obsolete form of retrieving mixin version, e.g. exec version, which returned an unstructured
// version string. It will be deprecated soon and is replaced by GetVersionMetadata.
func (p *FileSystem) GetVersion(m mixin.Metadata) (string, error) {
	r := mixin.NewRunner(m.Name, m.Dir, false)
	r.Command = "version"

	// Copy the existing context and tweak to pipe the output differently
	mixinVersion := &bytes.Buffer{}
	var mixinContext context.Context
	mixinContext = *p.Context
	mixinContext.Out = mixinVersion
	if !p.Debug {
		mixinContext.Err = ioutil.Discard
	}
	r.Context = &mixinContext

	err := r.Run()
	if err != nil {
		return "", err
	}

	return mixinVersion.String(), nil
}

// GetVersionMetadata is the new form of retrieving mixin version, e.g. exec version --output json, which returns
// a structured version string. It replaces GetVersion.
func (p *FileSystem) GetVersionMetadata(m mixin.Metadata) (*mixin.VersionInfo, error) {
	r := mixin.NewRunner(m.Name, m.Dir, false)
	r.Command = "version --output json"

	// Copy the existing context and tweak to pipe the output differently
	jsonB := &bytes.Buffer{}
	var mixinContext context.Context
	mixinContext = *p.Context
	mixinContext.Out = jsonB
	if !p.Debug {
		mixinContext.Err = ioutil.Discard
	}
	r.Context = &mixinContext

	err := r.Run()
	if err != nil {
		return nil, err
	}

	var response mixin.VersionInfo
	err = json.Unmarshal(jsonB.Bytes(), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
