package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/porter/version"
	"get.porter.sh/porter/pkg/printer"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
)

// HandshakeConfig is common handshake config between Porter and its plugins.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PORTER",
	MagicCookieValue: "bbc2dd71-def4-4311-906e-e98dc27208ce",
}

type PluginKey struct {
	Binary         string
	Interface      string
	Implementation string
	IsInternal     bool
}

// Implementaion stores implementation type (e.g. instance-storage) and its name (e.g. s3, mongo)
type Implementaion struct {
	Type string `json:"type"`
	Name string `json:"implementation"`
}

// PluginMetadata about an installed plugin.
type PluginMetadata struct {
	Name            string          `json:"name"`
	ClientPath      string          `json:"clientPath,omitempty"`
	Implementations []Implementaion `json:"implementations"`
	VersionInfo
}

// VersionInfo contains information from running the `version` command against the plugin.
type VersionInfo struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Author  string `json:"author,omitempty"`
}

type PluginProvider interface {
	List() ([]string, error)
	GetMetadata(string) (*PluginMetadata, error)
}

func NewFileSystem(config *config.Config) *fileSystem {
	return &fileSystem{
		Config: config,
	}
}

type fileSystem struct {
	*config.Config
}

func (fs *fileSystem) List() ([]string, error) {
	pluginsDir, err := fs.GetPluginsDir()
	if err != nil {
		return nil, err
	}

	files, err := fs.FileSystem.ReadDir(pluginsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list the contents of the plugins directory %q", pluginsDir)
	}

	plugins := make([]string, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			plugins = append(plugins, file.Name())
		}
	}

	return plugins, nil
}
func (fs *fileSystem) GetMetadata(pluginName string) (*PluginMetadata, error) {
	r := NewRunner(pluginName)

	// Copy the existing context and tweak to pipe the output differently
	jsonB := &bytes.Buffer{}

	mixinContext := *fs.Context
	mixinContext.Out = jsonB
	if !fs.Debug {
		mixinContext.Err = ioutil.Discard
	}
	r.Context = &mixinContext

	cmd := CommandOptions{Command: "version --output json"}
	err := r.Run(cmd)
	if err != nil {
		return nil, err
	}

	var mtd PluginMetadata
	err = json.Unmarshal(jsonB.Bytes(), &mtd)
	if err != nil {
		return nil, err
	}

	// make json.Marshal return `[]` instead of `nil` for not having implementations
	if len(mtd.Implementations) == 0 {
		mtd.Implementations = make([]Implementaion, 0)
	}
	return &mtd, nil
}

// PrintVersion writes plugin metadata to `ctx.Out` as plaintext or as json format based on `opts`
func PrintVersion(ctx *context.Context, opts version.Options, metadata PluginMetadata) error {
	switch opts.Format {
	case printer.FormatJson:
		return printer.PrintJson(ctx.Out, metadata)
	case printer.FormatPlaintext:
		authorship := ""
		if metadata.VersionInfo.Author != "" {
			authorship = " by " + metadata.VersionInfo.Author
		}
		_, err := fmt.Fprintf(ctx.Out, "%s %s (%s)%s\n", metadata.Name, metadata.VersionInfo.Version, metadata.VersionInfo.Commit, authorship)
		return err
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

func (k PluginKey) String() string {
	return fmt.Sprintf("%s.%s.%s", k.Interface, k.Binary, k.Implementation)
}

func ParsePluginKey(value string) (PluginKey, error) {
	var key PluginKey

	parts := strings.Split(value, ".")

	switch len(parts) {
	case 1:
		key.IsInternal = true
		key.Binary = "porter"
		key.Implementation = parts[0]
	case 2:
		key.Binary = parts[0]
		key.Implementation = parts[1]
	case 3:
		key.Interface = parts[0]
		key.Binary = parts[1]
		key.Implementation = parts[2]
	default:
		return PluginKey{}, errors.New("invalid plugin key %q, allowed format is [INTERFACE].BINARY.IMPLEMENTATION")
	}

	return key, nil
}
