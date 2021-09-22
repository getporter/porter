package smoke

import (
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

	// PorterHomeDir is the temp PORTER_HOME directory for the test.
	PorterHomeDir string

	// PorterPath is the path to the porter binary used for the test.
	PorterPath string

	// T is the test helper.
	T *testing.T
}

// NewTest sets up for an smoke test.
//
// Always defer Test.Teardown(), even when an error is returned.
func NewTest(t *testing.T) (Test, error) {
	var err error
	test := &Test{T: t}

	test.TestDir, err = ioutil.TempDir("", "porter-test")
	if err != nil {
		return *test, errors.Wrap(err, "could not create temp test directory")
	}

	err = test.createPorterHome()
	if err != nil {
		return *test, err
	}

	err = test.PorterE("help")
	if err != nil {
		return *test, err
	}

	err = test.PorterE("version")
	if err != nil {
		return *test, err
	}

	return *test, nil
}

// Run a porter command and fail the test if the command returns an error.
func (t Test) RequirePorter(args ...string) {
	err := t.PorterE(args...)
	require.NoError(t.T, err)
}

// Run a porter command, printing stderr when the command fails.
func (t Test) PorterE(args ...string) error {
	args = append(args, "--debug")
	p := shx.Command(t.PorterPath, args...).Stdout(nil)
	p.Cmd.Env = []string{"PORTER_HOME=" + t.PorterHomeDir}
	err := p.Run()
	return errors.Wrapf(err, "error running porter %s", args)
}

func (t Test) Teardown() {
	t.T.Log("Removing temp test PORTER_HOME")
	os.RemoveAll(t.PorterHomeDir)

	t.T.Log("Removing temp test directory")
	os.RemoveAll(t.TestDir)
}

// Create a test PORTER_HOME directory.
func (t *Test) createPorterHome() error {
	binDir, err := filepath.Abs("../../bin")
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
