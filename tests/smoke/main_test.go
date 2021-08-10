package smoke

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
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

	// RepoRoot is the root of the porter repository.
	// Useful for constructing paths that won't break when the test is moved.
	RepoRoot string

	// T is the test helper.
	T *testing.T
}

// NewTest sets up for a smoke test.
//
// Always defer Test.Teardown(), even when an error is returned.
func NewTest(t *testing.T) (Test, error) {
	os.Setenv(mg.VerboseEnv, "true")

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

	os.Setenv("PORTER_HOME", test.PorterHomeDir)

	err = test.PorterE("help")
	if err != nil {
		return *test, err
	}

	err = test.PorterE("version")
	if err != nil {
		return *test, err
	}

	return *test, test.startMongo()
}

func (t Test) startMongo() error {
	c := context.NewTestContext(t.T)
	conn, err := mongodb_docker.EnsureMongoIsRunning(c.Context, "porter-smoke-test-mongodb-plugin", "27017", "", "porter-smoke-test")
	defer conn.Close()
	if err != nil {
		return err
	}

	// Start with a fresh database
	err = conn.RemoveDatabase()
	return err
}

// Run a porter command and fail the test if the command returns an error.
func (t Test) RequirePorter(args ...string) {
	err := t.Porter(args...).RunV()
	require.NoError(t.T, err)
}

// Run a porter command, printing stderr when the command fails.
func (t Test) PorterE(args ...string) error {
	err := t.Porter(args...).RunE()
	return errors.Wrapf(err, "error running porter %s", args)
}

// Build a porter command, ready to be executed or further customized.
func (t Test) Porter(args ...string) shx.PreparedCommand {
	args = append(args, "--debug")
	return shx.Command(t.PorterPath, args...).
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
	t.RepoRoot = cxt.FindRepoRoot()

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

	cxt.CopyFile("testdata/config.toml", filepath.Join(t.PorterHomeDir, "config.toml"))

	return nil
}
