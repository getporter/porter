package porter

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
)

const (
	defaultEditor        = "vi"
	defaultEditorWindows = "notepad"
)

func editorCommand() string {
	editor := defaultEditor
	if runtime.GOOS == "windows" {
		editor = defaultEditorWindows
	}
	return editor
}

// RunEditor displays content to a user using an external text editor, like vi or notepad.
// The content is captured and returned during `Run()`
//
// The `EDITOR` environment variable is checked to find an editor.
// Failing that, use some sensible default depending on the operating system.
//
// This is useful for editing things like configuration files, especially those
// that might be stored on a remote server. For example: the content could be retrieved
// from the remote store, edited locally, then saved back.
func RunEditor(data []byte) ([]byte, error) {
	tempFile, err := ioutil.TempFile(os.TempDir(), "*")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(data)
	if err != nil {
		return nil, err
	}

	// close here without defer so the next command can grab the file
	tempFile.Close()

	// todo: ensure editor command with spaces works here
	// todo: other editors like Visual Studio Code don't seem to work with this, and don't block execution until the
	// editor has been saved like Notepad or vi do.
	cmd := exec.Command(editorCommand(), tempFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	data, err = ioutil.ReadFile(tempFile.Name())
	if err != nil {
		return nil, err
	}

	return data, nil
}
