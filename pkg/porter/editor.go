package porter

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	defaultEditor        = "vi"
	defaultEditorWindows = "notepad"
)

func editorArgs() []string {
	editor := defaultEditor
	if runtime.GOOS == "windows" {
		editor = defaultEditorWindows
	}
	if os.Getenv("EDITOR") != "" {
		editor = os.Getenv("EDITOR")
	}

	args := []string{}

	// any spaces need to be considered a separate argument passed to exec.
	// An example of where this would be needed is using "code.exe --wait"
	// for the EDITOR environment variable.
	if strings.ContainsAny(editor, "\"'\\") {
		args = strings.Split(editor, " ")
	} else {
		args = append(args, editor)
	}

	return args
}

// RunEditor displays content to a user using an external text editor, like vi or notepad.
// The content is captured and returned.
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

	args := editorArgs()
	args = append(args, tempFile.Name())
	cmd := exec.Command(args[0], args[1:]...)
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
