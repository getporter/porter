package manifest

import (
	"bytes"
	"io/ioutil"
	"sync"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"github.com/mikefarah/yq/v3/pkg/yqlib"
	"github.com/pkg/errors"
	"gopkg.in/op/go-logging.v1"
	"gopkg.in/yaml.v3"
)

var (
	// Only initialize the yq logging backend a single time, in a thread-safe way
	yqLogInit sync.Once
)

// Editor can modify the yaml in a Porter manifest.
type Editor struct {
	context  *context.Context
	manifest *Manifest
	yq       yqlib.YqLib
	node     *yaml.Node
}

func NewEditor(cxt *context.Context) *Editor {
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
		return len(data), errors.Wrapf(err, "could not parse manifest:\n%s", string(data))
	}

	return len(data), nil
}

func (e *Editor) ReadFile(src string) error {
	contents, err := e.context.FileSystem.ReadFile(src)
	if err != nil {
		return errors.Wrapf(err, "could not read the manifest at %q", src)
	}
	_, err = e.Read(contents)
	return errors.Wrapf(err, "could not parse the manifest at %q", src)
}

func (e *Editor) WriteFile(dest string) error {
	destFile, err := e.context.FileSystem.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "could not open destination manifest location %s", config.Name)
	}
	defer destFile.Close()

	// Encode the updated manifest to the proper location
	// yqlib.NewYamlEncoder takes: dest (io.Writer), indent spaces (int), colorized output (bool)
	var encoder = yqlib.NewYamlEncoder(destFile, 2, false)
	return errors.Wrapf(encoder.Encode(e.node), "unable to write the manifest to %s", dest)
}

func (e *Editor) SetValue(path string, value string) error {
	var valueParser = yqlib.NewValueParser()
	// valueParser.Parse takes: argument (string), custom tag (string),
	// custom style (string), anchor name (string), create alias (bool)
	var parsedValue = valueParser.Parse(value, "", "", "", false)
	cmd := yqlib.UpdateCommand{Command: "update", Path: path, Value: parsedValue, Overwrite: true}
	err := e.yq.Update(e.node, cmd, true)
	return errors.Wrapf(err, "could not update manifest path %q with value %q", path, value)
}
