package smoke

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/tests"
	"github.com/carolynvs/magex/shx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type Test struct {
	// TestDir is the temp directory created for the test.
	TestDir string

	// TestContext is a porter context for the filesystem.
	TestContext *context.TestContext

	// PorterHomeDir is the temp PORTER_HOME directory for the test.
	PorterHomeDir string

	// PorterPath is the path to the porter binary used for the test.
	PorterPath string

	// RepoRoot is the root of the porter repository.
	// Useful for constructing paths that won't break when the test is moved.
	RepoRoot string

	// T is the test helper.
	T *testing.T
}

// NewTest sets up for an smoke test.
//
// Always defer Test.Teardown(), even when an error is returned.
func NewTest(t *testing.T) (Test, error) {
	var err error
	test := &Test{T: t}

	test.TestContext = context.NewTestContext(t)
	test.TestContext.UseFilesystem()
	test.RepoRoot = test.TestContext.FindRepoRoot()

	test.TestDir, err = ioutil.TempDir("", "porter-test")
	if err != nil {
		return *test, errors.Wrap(err, "could not create temp test directory")
	}

	err = test.createPorterHome()
	if err != nil {
		return *test, err
	}

	os.Setenv("PORTER_HOME", test.PorterHomeDir)
	os.Setenv("PATH", test.PorterHomeDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	_, _, err = test.RunPorter("help")
	if err != nil {
		return *test, err
	}

	_, _, err = test.RunPorter("version")
	if err != nil {
		return *test, err
	}

	return *test, nil
}

// Run a porter command and fail the test if the command returns an error.
func (t Test) RequirePorter(args ...string) {
	_, _, err := t.RunPorter(args...)
	require.NoError(t.T, err)
}

// Run a porter command returning stderr when it fails
func (t Test) RunPorter(args ...string) (stdout string, combinedoutput string, err error) {
	t.T.Helper()
	// Copy stderr to stdout so we can return the "full" output printed to the console
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	output := &bytes.Buffer{}
	cmd := t.buildPorterCommand(args...)
	t.T.Log(cmd.String())
	ran, _, err := cmd.Stdout(io.MultiWriter(stdoutBuf, output)).Stderr(io.MultiWriter(stderrBuf, output)).Exec()
	if err != nil {
		if ran {
			err = errors.New(stderrBuf.String())
		}
		return stdoutBuf.String(), output.String(), err
	}
	t.T.Log(output.String())
	return stdoutBuf.String(), output.String(), nil
}

// Build a porter command, ready to be executed or further customized.
func (t Test) buildPorterCommand(args ...string) shx.PreparedCommand {
	args = append(args, "--debug")
	return shx.Command("porter", args...).
		Env("PORTER_HOME=" + t.PorterHomeDir)
}

func (t Test) Teardown() {
	t.T.Log("Removing temp test PORTER_HOME")
	os.RemoveAll(t.PorterHomeDir)

	t.T.Log("Removing temp test directory")
	os.RemoveAll(t.TestDir)
}

// Create a test PORTER_HOME directory.
func (t *Test) createPorterHome() error {
	var err error
	binDir := filepath.Join(t.RepoRoot, "bin")
	t.PorterHomeDir, err = ioutil.TempDir("", "porter")
	if err != nil {
		return errors.Wrap(err, "could not find the absolute path to bin/")
	}

	t.PorterHomeDir, err = ioutil.TempDir("", "porter")
	if err != nil {
		return errors.Wrap(err, "could not create temp PORTER_HOME directory")
	}

	cxt := context.NewTestContext(t.T)
	cxt.UseFilesystem()

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	t.PorterPath = filepath.Join(t.PorterHomeDir, "porter"+ext)
	err = cxt.CopyFile(filepath.Join(binDir, "porter"+ext), t.PorterPath)
	if err != nil {
		return errors.Wrap(err, "could not copy porter binary into test PORTER_HOME")
	}

	err = cxt.CopyDirectory(filepath.Join(binDir, "runtimes"), t.PorterHomeDir, true)
	if err != nil {
		return errors.Wrap(err, "could not copy runtimes/ into test PORTER_HOME")
	}

	err = cxt.CopyDirectory(filepath.Join(binDir, "mixins"), t.PorterHomeDir, true)
	if err != nil {
		return errors.Wrap(err, "could not copy mixins/ into test PORTER_HOME")
	}

	return nil
}

// RequireFileMode checks that all files in the specified path match the specifed
// file mode. Uses a glob pattern to match.
func (t *Test) RequireFileMode(path string, mode os.FileMode) {
	if !tests.AssertDirectoryPermissionsEqual(t.T, path, mode) {
		t.T.FailNow()
	}
}
