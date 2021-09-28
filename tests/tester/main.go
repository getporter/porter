package tester

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"
	"get.porter.sh/porter/tests"
	"github.com/carolynvs/magex/shx"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type Tester struct {
	originalPwd string

	// unique database name assigned to the test
	dbName string

	// TestContext is a porter context for the filesystem.
	TestContext *context.TestContext

	// TestDir is the temp directory created for the test.
	TestDir string

	// PorterHomeDir is the temp PORTER_HOME directory for the test.
	PorterHomeDir string

	// RepoRoot is the root of the porter repository.
	// Useful for constructing paths that won't break when the test is moved.
	RepoRoot string

	// T is the test helper.
	T *testing.T
}

// NewTest sets up for a smoke test.
//
// Always defer Tester.Teardown(), even when an error is returned.
func NewTest(t *testing.T) (Tester, error) {
	var err error
	pwd, _ := os.Getwd()
	test := &Tester{T: t, originalPwd: pwd}

	test.TestContext = context.NewTestContext(t)
	test.TestContext.UseFilesystem()
	test.RepoRoot = test.TestContext.FindRepoRoot()

	test.TestDir, err = ioutil.TempDir("", "porter-test")
	if err != nil {
		return *test, errors.Wrap(err, "could not create temp test directory")
	}

	test.dbName = tests.GenerateDatabaseName(t.Name())

	err = test.createPorterHome()
	if err != nil {
		return *test, err
	}

	os.Setenv("PORTER_HOME", test.PorterHomeDir)
	os.Setenv("PATH", test.PorterHomeDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return *test, test.startMongo()
}

// CurrentNamespace configured in config.toml
func (t Tester) CurrentNamespace() string {
	return "dev"
}

func (t Tester) startMongo() error {
	conn, err := mongodb_docker.EnsureMongoIsRunning(t.TestContext.Context, "porter-smoke-test-mongodb-plugin", "27017", "", t.dbName, 2)
	defer conn.Close()
	if err != nil {
		return err
	}

	// Start with a fresh database
	err = conn.RemoveDatabase()
	return err
}

// Run a porter command and fail the test if the command returns an error.
func (t Tester) RequirePorter(args ...string) (string, string) {
	t.T.Helper()
	stdout, output, err := t.RunPorter(args...)
	require.NoError(t.T, err)
	return stdout, output
}

// Run a porter command returning stderr when it fails
func (t Tester) RunPorter(args ...string) (stdout string, combinedoutput string, err error) {
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
	return stdoutBuf.String(), output.String(), nil
}

// Build a porter command, ready to be executed or further customized.
func (t Tester) buildPorterCommand(args ...string) shx.PreparedCommand {
	args = append(args, "--debug")
	return shx.Command("porter", args...).
		Env("PORTER_HOME=" + t.PorterHomeDir)
}

func (t Tester) Teardown() {
	t.T.Log("Removing temp test PORTER_HOME")
	os.RemoveAll(t.PorterHomeDir)

	t.T.Log("Removing temp test directory")
	os.RemoveAll(t.TestDir)

	// Reset the current directory for the next test
	t.Chdir(t.originalPwd)
}

// Create a test PORTER_HOME directory.
func (t *Tester) createPorterHome() error {
	var err error
	binDir := filepath.Join(t.RepoRoot, "bin")
	t.PorterHomeDir, err = ioutil.TempDir("", "porter")
	if err != nil {
		return errors.Wrap(err, "could not create temp PORTER_HOME directory")
	}

	require.NoError(t.T, shx.Copy(filepath.Join(binDir, "porter*"), t.PorterHomeDir),
		"could not copy porter binaries into test PORTER_HOME")
	require.NoError(t.T, shx.Copy(filepath.Join(binDir, "runtimes"), t.PorterHomeDir, shx.CopyRecursive),
		"could not copy runtimes/ into test PORTER_HOME")
	require.NoError(t.T, shx.Copy(filepath.Join(binDir, "mixins"), t.PorterHomeDir, shx.CopyRecursive),
		"could not copy mixins/ into test PORTER_HOME")

	// Write out a config file with a unique database name set
	cfgD, err := ioutil.ReadFile(filepath.Join(t.RepoRoot, "tests/testdata/config/config.toml"))
	require.NoError(t.T, err)
	cfgD = bytes.Replace(cfgD, []byte("porter-test"), []byte(t.dbName), 1)
	err = ioutil.WriteFile(filepath.Join(t.PorterHomeDir, "config.toml"), cfgD, 0700)
	require.NoError(t.T, err, "could not copy config.toml into test PORTER_HOME")

	return nil
}

func (t Tester) Chdir(dir string) {
	t.TestContext.Chdir(dir)
	os.Chdir(dir)
}
