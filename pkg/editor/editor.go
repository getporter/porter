package editor

import (
	"os"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/context"
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
	*context.Context
	data         []byte
	tempFilename string
}

// New returns a new editor with the temp filename and data provided.
func New(context *context.Context, tempFilename string, data []byte) *Editor {
	return &Editor{
		Context:      context,
		tempFilename: tempFilename,
		data:         data,
	}
}

func editorArgs() []string {
	editor := defaultEditor
	if os.Getenv("EDITOR") != "" {
		editor = os.Getenv("EDITOR")
	}

	args := []string{}
	// split any spaces in the editor command into separate arguments, otherwise
	// exec will treat the entire EDITOR value as one filename.
	// for example: EDITOR set as "/tmp/myeditor --wait" will be split into "/tmp/myeditor" and
	// "--wait" as the second argument.
	// another example: "C:\Program Files\Visual Studio Code\Code.exe --wait" will
	// be split into "C:\Program", "Files\Visual", "Studio", "Code\Code.exe", "--wait"
	// but this works correctly with exec.Command() despite looking a bit strange.
	if strings.ContainsAny(editor, " ") {
		args = strings.Split(editor, " ")
	} else {
		args = append(args, editor)
	}

	return args
}

// Run opens the editor, displaying the content through a temporary file.
// The content is returned once the editor closes.
func (e *Editor) Run() ([]byte, error) {
	tempFile, err := e.FileSystem.OpenFile(filepath.Join(os.TempDir(), e.tempFilename), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	defer e.FileSystem.Remove(tempFile.Name())

	_, err = tempFile.Write(e.data)
	if err != nil {
		return nil, err
	}

	// close here without defer so cmd can grab the file
	tempFile.Close()

	args := editorArgs()
	args = append(args, tempFile.Name())
	cmd := e.NewCommand(args[0], args[1:]...)
	cmd.Stdout = e.Out
	cmd.Stderr = e.Err
	cmd.Stdin = e.In
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	data, err := e.FileSystem.ReadFile(tempFile.Name())
	if err != nil {
		return nil, err
	}

	return data, nil
}
