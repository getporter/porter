package editor

import (
	"fmt"
	"os"
	"path/filepath"

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
	contents     []byte
	tempFilename string
}

// New returns a new Editor with the temp filename and contents provided.
func New(context *context.Context, tempFilename string, contents []byte) *Editor {
	return &Editor{
		Context:      context,
		tempFilename: tempFilename,
		contents:     contents,
	}
}

func editorArgs(filename string) []string {
	shell := defaultShell
	if os.Getenv("SHELL") != "" {
		shell = os.Getenv("SHELL")
	}
	editor := defaultEditor
	if os.Getenv("EDITOR") != "" {
		editor = os.Getenv("EDITOR")
	}

	// Example of what will be run:
	// on *nix: sh -c "vi /tmp/test.txt"
	// on windows: cmd /C "C:\Program Files\Visual Studio Code\Code.exe --wait C:\somefile.txt"
	//
	// Pass the editor command to the shell so we don't have to parse the command ourselves.
	// Passing the editor command that could possibly have an argument (e.g. --wait for VSCode) to the
	// shell means we don't have to parse this ourselves, like splitting on spaces.
	return []string{shell, shellCommandFlag, fmt.Sprintf("%s %s", editor, filename)}
}

// Run opens the editor, displaying the contents through a temporary file.
// The content is returned once the editor closes.
func (e *Editor) Run() ([]byte, error) {
	tempFile, err := e.FileSystem.OpenFile(filepath.Join(os.TempDir(), e.tempFilename), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return nil, err
	}
	defer e.FileSystem.Remove(tempFile.Name())

	_, err = tempFile.Write(e.contents)
	if err != nil {
		return nil, err
	}

	// close here without defer so cmd can grab the file
	tempFile.Close()

	args := editorArgs(tempFile.Name())
	cmd := e.NewCommand(args[0], args[1:]...)
	cmd.Stdout = e.Out
	cmd.Stderr = e.Err
	cmd.Stdin = e.In
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	contents, err := e.FileSystem.ReadFile(tempFile.Name())
	if err != nil {
		return nil, err
	}

	return contents, nil
}
