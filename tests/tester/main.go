package tester

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
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
	TestContext *portercontext.TestContext

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
// Always defer Tester.Close(), even when an error is returned.
func NewTest(t *testing.T) (Tester, error) {
	return NewTestWithConfig(t, "")
}

// NewTestWithConfig sets up for a smoke test using the specified
// Porter config file. The path should be either be absolute, or
// relative to the repository root.
//
// Always defer Tester.Close(), even when an error is returned.
func NewTestWithConfig(t *testing.T, configFilePath string) (Tester, error) {
	var err error
	pwd, _ := os.Getwd()
	test := &Tester{T: t, originalPwd: pwd}

	test.TestContext = portercontext.NewTestContext(t)
	test.TestContext.UseFilesystem()
	test.RepoRoot = test.TestContext.FindRepoRoot()

	test.TestDir, err = ioutil.TempDir("", "porter-test")
	if err != nil {
		return *test, errors.Wrap(err, "could not create temp test directory")
	}

	test.dbName = tests.GenerateDatabaseName(t.Name())

	err = test.createPorterHome(configFilePath)
	if err != nil {
		return *test, err
	}

	os.Setenv("PORTER_HOME", test.PorterHomeDir)
	os.Setenv("PATH", test.PorterHomeDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return *test, test.startMongo(context.Background())
}

// CurrentNamespace configured in Porter's config file
func (t Tester) CurrentNamespace() string {
	return "dev"
}

func (t Tester) startMongo(ctx context.Context) error {
	conn, err := mongodb_docker.EnsureMongoIsRunning(ctx,
		t.TestContext.Context,
		"porter-smoke-test-mongodb-plugin",
		"27017",
		"",
		t.dbName,
		10,
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Start with a fresh database
	err = conn.RemoveDatabase(ctx)
	return err
}

// Run a porter command and fail the test if the command returns an error.
func (t Tester) RequirePorter(args ...string) (string, string) {
	t.T.Helper()
	stdout, output, err := t.RunPorter(args...)
	require.NoError(t.T, err)
	return stdout, output
}

// RunPorter executes a porter command returning stderr when it fails.
func (t Tester) RunPorter(args ...string) (stdout string, combinedoutput string, err error) {
	t.T.Helper()
	return t.RunPorterWith(func(cmd *shx.PreparedCommand) {
		cmd.Args(args...)
	})
}

// RunPorterWith works like RunPorter, but you can customize the command before it's run.
func (t Tester) RunPorterWith(opts ...func(*shx.PreparedCommand)) (stdout string, combinedoutput string, err error) {
	t.T.Helper()

	cmd := t.buildPorterCommand(opts...)

	// Copy stderr to stdout so we can return the "full" output printed to the console
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	output := &bytes.Buffer{}
	cmd.Stdout(io.MultiWriter(stdoutBuf, output)).Stderr(io.MultiWriter(stderrBuf, output))

	t.T.Log(cmd.String())
	ran, _, err := cmd.Exec()
	if err != nil {
		if ran {
			err = errors.Wrap(err, stderrBuf.String())
		}
		return stdoutBuf.String(), output.String(), err
	}
	return stdoutBuf.String(), output.String(), nil
}

// Build a porter command, ready to be executed or further customized.
func (t Tester) buildPorterCommand(opts ...func(*shx.PreparedCommand)) shx.PreparedCommand {
	cmd := shx.Command("porter", "--debug").
		Env("PORTER_HOME="+t.PorterHomeDir, "PORTER_TEST_DB_NAME="+t.dbName)
	for _, opt := range opts {
		opt(&cmd)
	}
	return cmd
}

func (t Tester) Close() {
	t.T.Log("Removing temp test PORTER_HOME")
	os.RemoveAll(t.PorterHomeDir)

	t.T.Log("Removing temp test directory")
	os.RemoveAll(t.TestDir)

	// Reset the current directory for the next test
	t.Chdir(t.originalPwd)
}

// Create a test PORTER_HOME directory with the optional config file.
// The config file path should be specified relative to the repository root
func (t *Tester) createPorterHome(configFilePath string) error {
	if configFilePath == "" {
		configFilePath = "tests/testdata/config/config.yaml"
	}
	if !filepath.IsAbs(configFilePath) {
		configFilePath = filepath.Join(t.RepoRoot, configFilePath)
	}

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
	require.NoError(t.T, shx.Copy(configFilePath, filepath.Join(t.PorterHomeDir, "config"+filepath.Ext(configFilePath))),
		"error copying config file to PORTER_HOME")

	return nil
}

func (t Tester) Chdir(dir string) {
	t.TestContext.Chdir(dir)
	os.Chdir(dir)
}
