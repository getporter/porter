package tester

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/storage/plugins/mongodb_docker"
	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/require"
	"github.com/uwu-tools/magex/shx"
)

// mongoBootstrapMu serializes the "does the shared mongo container exist,
// if not create it" check so that parallel tests don't race to `docker run`
// the same fixed container name. Each test still creates/removes its own
// database inside that shared container concurrently.
var mongoBootstrapMu sync.Mutex

// syncBuffer is a bytes.Buffer safe for concurrent writes. Needed anywhere a
// buffer is shared between a command's stdout and stderr, since exec.Cmd
// copies each of those streams in its own concurrent goroutine.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

type Tester struct {
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

	// extraEnv holds additional environment variables set via SetEnv, applied
	// to every porter subprocess this Tester runs. It's a plain map (not a
	// pointer) because map values already have reference semantics, so
	// mutations through a value-receiver method are visible to every copy of
	// this Tester.
	extraEnv map[string]string

	// T is the test helper.
	T *testing.T
}

// SetEnv sets an environment variable that is passed to every porter command
// this Tester runs afterward. Unlike os.Setenv, this only affects this
// Tester's own subprocesses, so it's safe to use from parallel tests.
func (t Tester) SetEnv(key, value string) {
	t.extraEnv[key] = value
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
	test := &Tester{T: t, extraEnv: make(map[string]string)}

	test.TestContext = portercontext.NewTestContext(t)
	test.TestContext.UseFilesystem()
	test.RepoRoot = test.TestContext.FindRepoRoot()

	test.TestDir, err = os.MkdirTemp("", "porter-test")
	if err != nil {
		return *test, fmt.Errorf("could not create temp test directory: %w", err)
	}

	test.dbName = tests.GenerateDatabaseName(t.Name())

	err = test.createPorterHome(configFilePath)
	if err != nil {
		return *test, err
	}

	return *test, test.startMongo(context.Background())
}

// CurrentNamespace configured in Porter's config file
func (t Tester) CurrentNamespace() string {
	return "dev"
}

func (t Tester) startMongo(ctx context.Context) error {
	mongoBootstrapMu.Lock()
	conn, err := mongodb_docker.EnsureMongoIsRunning(ctx,
		t.TestContext.Context,
		"porter-smoke-test-mongodb-plugin",
		"27017",
		"",
		t.dbName,
		10,
	)
	mongoBootstrapMu.Unlock()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Start with a fresh database
	err = conn.RemoveDatabase(ctx)
	return err
}

// Run a porter command and fail the test if the command returns an error.
func (t Tester) RequirePorter(args ...string) (stdout string, combinedoutput string) {
	t.T.Helper()
	stdout, combinedoutput, err := t.RunPorter(args...)
	if err != nil {
		t.T.Logf("failed to run porter %s", strings.Join(args, " "))
		t.T.Log(combinedoutput)
	}
	require.NoError(t.T, err)
	return stdout, combinedoutput
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

	// Copy stderr to stdout so we can return the "full" output printed to the console.
	// exec.Cmd copies stdout and stderr concurrently in separate goroutines, so the
	// buffer they both write into (output) needs its own synchronization; stdoutBuf
	// and stderrBuf are each only ever written by a single goroutine and don't.
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	output := &syncBuffer{}
	cmd.Stdout(io.MultiWriter(stdoutBuf, output)).Stderr(io.MultiWriter(stderrBuf, output))

	t.T.Log(cmd.String())
	ran, _, err := cmd.Exec()
	if err != nil {
		if ran {
			err = fmt.Errorf("%s: %w", stderrBuf.String(), err)
		}
		return stdoutBuf.String(), output.String(), err
	}
	return stdoutBuf.String(), output.String(), nil
}

// Build a porter command, ready to be executed or further customized.
func (t Tester) buildPorterCommand(opts ...func(*shx.PreparedCommand)) shx.PreparedCommand {
	debugCmdPrefix := os.Getenv("PORTER_RUN_IN_DEBUGGER")

	configureCommand := func(cmd shx.PreparedCommand) {
		cmd.In(t.TestContext.Getwd())
		cmd.Env(
			"PORTER_HOME="+t.PorterHomeDir,
			"PORTER_TEST_DB_NAME="+t.dbName,
			"PORTER_VERBOSITY=debug",
		)
		for k, v := range t.extraEnv {
			cmd.Env(k + "=" + v)
		}
		for _, opt := range opts {
			opt(&cmd)
		}
	}

	// Invoke the copied porter binary by its absolute path so that we never
	// depend on (or race on) the real process PATH/LookPath.
	cmd := shx.Command(filepath.Join(t.PorterHomeDir, "porter"))
	// Keep Args[0]/display output as "porter" so PORTER_RUN_IN_DEBUGGER prefix
	// matching below, and log output, still read as "porter ...".
	cmd.Cmd.Args[0] = "porter"
	configureCommand(cmd)

	prettyCmd := cmd.String()
	if debugCmdPrefix != "" && strings.HasPrefix(prettyCmd, debugCmdPrefix) {
		port := os.Getenv("PORTER_DEBUGGER_PORT")
		if port == "" {
			port = "55942"
		}
		porterPath := filepath.Join(t.RepoRoot, "bin/porter")
		cmd = shx.Command("dlv", "exec", porterPath, "--listen=:"+port, "--headless=true", "--api-version=2", "--accept-multiclient", "--")
		configureCommand(cmd)
	}

	return cmd
}

func (t Tester) Close() {
	t.T.Log("Removing temp test PORTER_HOME")
	os.RemoveAll(t.PorterHomeDir)

	t.T.Log("Removing temp test directory")
	os.RemoveAll(t.TestDir)
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
	t.PorterHomeDir, err = os.MkdirTemp("", "porter")
	if err != nil {
		return fmt.Errorf("could not create temp PORTER_HOME directory: %w", err)
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
}
