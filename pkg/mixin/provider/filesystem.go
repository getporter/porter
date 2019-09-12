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

func (fs *FileSystem) List() ([]mixin.Metadata, error) {
	mixinsDir, err := fs.GetMixinsDir()
	if err != nil {
		return nil, err
	}

	files, err := fs.FileSystem.ReadDir(mixinsDir)
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

func (fs *FileSystem) GetSchema(m mixin.Metadata) (string, error) {
	r := NewRunner(m.Name, m.Dir, false)

	// Copy the existing context and tweak to pipe the output differently
	mixinSchema := &bytes.Buffer{}
	var mixinContext context.Context
	mixinContext = *fs.Context
	mixinContext.Out = mixinSchema
	if !fs.Debug {
		mixinContext.Err = ioutil.Discard
	}
	r.Context = &mixinContext

	cmd := mixin.CommandOptions{Command: "schema"}
	err := r.Run(cmd)
	if err != nil {
		return "", err
	}

	return mixinSchema.String(), nil
}

// GetVersion is the obsolete form of retrieving mixin version, e.g. exec version, which returned an unstructured
// version string. It will be deprecated soon and is replaced by GetVersionMetadata.
func (fs *FileSystem) GetVersion(m mixin.Metadata) (string, error) {
	r := NewRunner(m.Name, m.Dir, false)

	// Copy the existing context and tweak to pipe the output differently
	mixinVersion := &bytes.Buffer{}
	var mixinContext context.Context
	mixinContext = *fs.Context
	mixinContext.Out = mixinVersion
	if !fs.Debug {
		mixinContext.Err = ioutil.Discard
	}
	r.Context = &mixinContext

	cmd := mixin.CommandOptions{Command: "version"}
	err := r.Run(cmd)
	if err != nil {
		return "", err
	}

	return mixinVersion.String(), nil
}

// GetVersionMetadata is the new form of retrieving mixin version, e.g. exec version --output json, which returns
// a structured version string. It replaces GetVersion.
func (fs *FileSystem) GetVersionMetadata(m mixin.Metadata) (*mixin.VersionInfo, error) {
	r := NewRunner(m.Name, m.Dir, false)

	// Copy the existing context and tweak to pipe the output differently
	jsonB := &bytes.Buffer{}
	var mixinContext context.Context
	mixinContext = *fs.Context
	mixinContext.Out = jsonB
	if !fs.Debug {
		mixinContext.Err = ioutil.Discard
	}
	r.Context = &mixinContext

	cmd := mixin.CommandOptions{Command: "version --output json"}
	err := r.Run(cmd)
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

func (fs *FileSystem) Run(mixinContext *context.Context, mixinName string, commandOpts mixin.CommandOptions) error {
	mixinDir, err := fs.GetMixinDir(mixinName)
	if err != nil {
		return err
	}

	r := NewRunner(mixinName, mixinDir, commandOpts.Runtime)
	r.Context = mixinContext

	err = r.Validate()
	if err != nil {
		return err
	}

	return r.Run(commandOpts)
}
