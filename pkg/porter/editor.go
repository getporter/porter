package porter

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"get.porter.sh/porter/pkg/context"
)

const (
	defaultEditor        = "vi"
	defaultEditorWindows = "notepad"
)

// Editor displays content to a user using an external text editor, like vi or notepad.
// The content is captured and returned.
//
// The `EDITOR` environment variable is checked to find an editor.
// Failing that, use some sensible default depending on the operating system.
//
// This is useful for editing things like configuration files, especially those
// that might be stored on a remote server. For example: the content could be retrieved
// from the remote store, edited locally, then saved back.
type Editor struct {
	Data     []byte
	Filename string
	Context  *context.Context
}

func editorArgs() []string {
	editor := defaultEditor
	if runtime.GOOS == "windows" {
		editor = defaultEditorWindows
	}
	if os.Getenv("EDITOR") != "" {
		editor = os.Getenv("EDITOR")
	}

	args := []string{}
	if strings.ContainsAny(editor, "\"'\\") {
		args = strings.Split(editor, " ")
	} else {
		args = append(args, editor)
	}

	return args
}

// Run opens the editor, displaying the content through a temporary file.
// The content is returned once the editor closes.
func (e *Editor) Run() ([]byte, error) {
	tempFile, err := os.OpenFile(filepath.Join(os.TempDir(), e.Filename), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(e.Data)
	if err != nil {
		return nil, err
	}

	// close here without defer so cmd can grab the file
	tempFile.Close()

	args := editorArgs()
	args = append(args, tempFile.Name())
	cmd := e.Context.NewCommand(args[0], args[1:]...)
	cmd.Stdout = e.Context.Out
	cmd.Stderr = e.Context.Err
	cmd.Stdin = e.Context.In
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		return nil, err
	}

	return data, nil
}
