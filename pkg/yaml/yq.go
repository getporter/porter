package yaml

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/portercontext"
	"github.com/mikefarah/yq/v3/pkg/yqlib"
	"gopkg.in/op/go-logging.v1"
	"gopkg.in/yaml.v3"
)

var (
	// Only initialize the yq logging backend a single time, in a thread-safe way
	yqLogInit sync.Once
)

// Editor can modify the yaml in a Porter manifest.
type Editor struct {
	context *portercontext.Context
	yq      yqlib.YqLib
	node    *yaml.Node
}

func NewEditor(cxt *portercontext.Context) *Editor {
	e := &Editor{
		context: cxt,
	}
	yqLogInit.Do(e.suppressYqLogging)
	return e
}

// Hide yq log statements
func (e *Editor) suppressYqLogging() {
	// TODO: We could improve the logging and yq libs to be parallel friendly
	// This turns off all logging from yq
	// yq doesn't return errors from its api, and logs them instead (awkward)
	// The yq logger is global, unless we seriously edit yq, we can't separate
	// logging for once instance of a yq run from any parallel runs. It just
	// seemed better to turn it off.
	// yq has moved to v4 which is a very large api change, so it would be
	// a lot of work though.

	// The yq lib that we use makes frequent calls to a logger that are by default
	// printed directly to stderr
	var backend = logging.AddModuleLevel(logging.NewLogBackend(ioutil.Discard, "", 0))
	backend.SetLevel(logging.ERROR, "yq")
	logging.SetBackend(backend)
}

func (e *Editor) Read(data []byte) (n int, err error) {
	e.yq = yqlib.NewYqLib()
	e.node = &yaml.Node{}

	var decoder = yaml.NewDecoder(bytes.NewReader(data))
	err = decoder.Decode(e.node)
	if err != nil {
		return len(data), fmt.Errorf("could not parse manifest:\n%s: %w", string(data), err)
	}

	return len(data), nil
}

func (e *Editor) ReadFile(src string) error {
	contents, err := e.context.FileSystem.ReadFile(src)
	if err != nil {
		return fmt.Errorf("could not read the manifest at %q: %w", src, err)
	}
	_, err = e.Read(contents)
	if err != nil {
		return fmt.Errorf("could not parse the manifest at %q: %w", src, err)
	}

	return nil
}

func (e *Editor) WriteFile(dest string) error {
	destFile, err := e.context.FileSystem.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, pkg.FileModeWritable)
	if err != nil {
		return fmt.Errorf("could not open destination manifest location %s: %w", config.Name, err)
	}
	defer destFile.Close()

	// Encode the updated manifest to the proper location
	// yqlib.NewYamlEncoder takes: dest (io.Writer), indent spaces (int), colorized output (bool)
	var encoder = yqlib.NewYamlEncoder(destFile, 2, false)
	err = encoder.Encode(e.node)
	if err != nil {
		return fmt.Errorf("unable to write the manifest to %s: %w", dest, err)
	}

	return nil
}

func (e *Editor) SetValue(path string, value string) error {
	var valueParser = yqlib.NewValueParser()
	// valueParser.Parse takes: argument (string), custom tag (string),
	// custom style (string), anchor name (string), create alias (bool)
	var parsedValue = valueParser.Parse(value, "", "", "", false)
	cmd := yqlib.UpdateCommand{Command: "update", Path: path, Value: parsedValue, Overwrite: true}
	err := e.yq.Update(e.node, cmd, true)
	if err != nil {
		return fmt.Errorf("could not update path %q with value %q: %w", path, value, err)
	}

	return nil
}

func (e *Editor) DeleteNode(path string) error {
	cmd := yqlib.UpdateCommand{Command: "delete", Path: path}
	err := e.yq.Update(e.node, cmd, true)
	if err != nil {
		return fmt.Errorf("could not delete path %q: %w", path, err)
	}

	return nil
}

// GetNode evaluates the specified yaml path to a single node.
// Returns an error if a node isn't found, or more than one is found.
func (e *Editor) GetNode(path string) (*yaml.Node, error) {
	results, err := e.yq.Get(e.node, path)
	if err != nil {
		return nil, err
	}

	switch len(results) {
	case 0:
		return nil, fmt.Errorf("no matching nodes found for %s", path)
	case 1:
		return results[0].Node, nil
	default:
		return nil, fmt.Errorf("multiple nodes matched the path %s", path)
	}
}

// WalkNodes executes f for all yaml nodes found in path.
// If an error is returned from f, the WalkNodes function will return the error and stop interating through
// the rest of the nodes.
func (e *Editor) WalkNodes(ctx context.Context, path string, f func(ctx context.Context, nc *yqlib.NodeContext) error) error {
	nodes, err := e.yq.Get(e.node, path)
	if err != nil {
		return fmt.Errorf("failed to find nodes with path %s: %w", path, err)
	}

	for _, node := range nodes {
		if node.Node.IsZero() {
			continue
		}
		if err := f(ctx, node); err != nil {
			return err
		}

	}

	return nil
}
