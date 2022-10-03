package portercontext

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/test"
	"github.com/carolynvs/aferox"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type TestContext struct {
	*Context

	cleanupDirs []string
	capturedErr *bytes.Buffer
	capturedOut *bytes.Buffer
	T           *testing.T
}

// NewTestContext initializes a configuration suitable for testing, with the
// output buffered, and an in-memory file system, using the specified
// environment variables.
func NewTestContext(t *testing.T) *TestContext {
	// Provide a way for tests to provide and capture stdin and stdout
	// Copy output to the test log simultaneously, use go test -v to see the output
	err := &bytes.Buffer{}
	aggErr := io.MultiWriter(err, test.Logger{T: t})
	out := &bytes.Buffer{}
	aggOut := io.MultiWriter(out, test.Logger{T: t})

	innerContext := New()
	innerContext.correlationId = "0"
	innerContext.timestampLogs = false
	innerContext.environ = getEnviron()
	innerContext.FileSystem = aferox.NewAferox("/", afero.NewMemMapFs())
	innerContext.In = &bytes.Buffer{}
	innerContext.Out = aggOut
	innerContext.Err = aggErr
	innerContext.ConfigureLogging(context.Background(), LogConfiguration{
		LogLevel:  zapcore.DebugLevel,
		Verbosity: zapcore.DebugLevel,
	})
	innerContext.PlugInDebugContext = &PluginDebugContext{
		DebuggerPort:           "2735",
		RunPlugInInDebugger:    "",
		PlugInWorkingDirectory: "",
	}

	c := &TestContext{
		Context:     innerContext,
		capturedOut: out,
		capturedErr: err,
		T:           t,
	}

	c.NewCommand = c.NewTestCommand

	return c
}

func (c *TestContext) NewTestCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	testArgs := append([]string{name}, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], testArgs...)
	cmd.Dir = c.Getwd()

	cmd.Env = []string{
		fmt.Sprintf("%s=true", test.MockedCommandEnv),
	}
	if val, ok := c.LookupEnv(test.ExpectedCommandEnv); ok {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", test.ExpectedCommandEnv, val))
	}
	if val, ok := c.LookupEnv(test.ExpectedCommandExitCodeEnv); ok {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", test.ExpectedCommandExitCodeEnv, val))
	}
	if val, ok := c.LookupEnv(test.ExpectedCommandOutputEnv); ok {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", test.ExpectedCommandOutputEnv, val))
	}
	if val, ok := c.LookupEnv(test.ExpectedCommandErrorEnv); ok {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", test.ExpectedCommandErrorEnv, val))
	}
	return cmd
}

func (c *TestContext) GetTestDefinitionDirectory() string {
	for i := 0; true; i++ {
		_, filename, _, ok := runtime.Caller(i)
		if !ok {
			c.T.Fatal("could not determine calling test directory")
		}
		if strings.HasSuffix(filename, "_test.go") {
			return filepath.Dir(filename)
		}
	}
	return ""
}

// UseFilesystem has porter's context use the OS filesystem instead of an in-memory filesystem
// Returns the test directory, and the temp porter home directory.
func (c *TestContext) UseFilesystem() (testDir string, homeDir string) {
	homeDir, err := ioutil.TempDir("", "porter-test")
	require.NoError(c.T, err)
	c.cleanupDirs = append(c.cleanupDirs, homeDir)

	testDir = c.GetTestDefinitionDirectory()
	c.FileSystem = aferox.NewAferox(testDir, afero.NewOsFs())
	c.defaultNewCommand()
	c.DisableUmask()

	return testDir, homeDir
}

func (c *TestContext) AddCleanupDir(dir string) {
	c.cleanupDirs = append(c.cleanupDirs, dir)
}

func (c *TestContext) Close() {
	for _, dir := range c.cleanupDirs {
		c.FileSystem.RemoveAll(dir)
	}
}

// AddTestFileFromRoot should be used when the testfile you are referencing is in a different directory than the test.
func (c *TestContext) AddTestFileFromRoot(src, dest string) []byte {
	pathFromRoot := filepath.Join(c.FindRepoRoot(), src)
	return c.AddTestFile(pathFromRoot, dest)
}

// AddTestFile adds a test file where the filepath is relative to the test directory.
// mode is optional and only the first one passed is used.
func (c *TestContext) AddTestFile(src, dest string, mode ...os.FileMode) []byte {
	if strings.Contains(src, "..") {
		c.T.Fatal(errors.New("use AddTestFileFromRoot when referencing a test file in a different directory than the test"))
	}

	data, err := ioutil.ReadFile(src)
	if err != nil {
		c.T.Fatal(fmt.Errorf("error reading file %s from host filesystem: %w", src, err))
	}

	var perms os.FileMode
	if len(mode) == 0 {
		ext := filepath.Ext(dest)
		if ext == ".sh" || ext == "" {
			perms = pkg.FileModeExecutable
		} else {
			perms = pkg.FileModeWritable
		}
	} else {
		perms = mode[0]
	}

	err = c.FileSystem.WriteFile(dest, data, perms)
	if err != nil {
		c.T.Fatal(fmt.Errorf("error writing file %s to test filesystem: %w", dest, err))
	}

	return data
}

func (c *TestContext) AddTestFileContents(file []byte, dest string) error {
	return c.FileSystem.WriteFile(dest, file, pkg.FileModeWritable)
}

// Use this when the directory you are referencing is in a different directory than the test.
func (c *TestContext) AddTestDirectoryFromRoot(srcDir, destDir string) {
	pathFromRoot := filepath.Join(c.FindRepoRoot(), srcDir)
	c.AddTestDirectory(pathFromRoot, destDir)
}

// AddTestDirectory adds a test directory where the filepath is relative to the test directory.
// mode is optional and should only be specified once
func (c *TestContext) AddTestDirectory(srcDir, destDir string, mode ...os.FileMode) {
	if strings.Contains(srcDir, "..") {
		c.T.Fatal(errors.New("use AddTestDirectoryFromRoot when referencing a test directory in a different directory than the test"))
	}

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root src directory
		if path == srcDir {
			return nil
		}

		// Translate the path from the src to the final destination
		dest := filepath.Join(destDir, strings.TrimPrefix(path, srcDir))

		if info.IsDir() {
			return c.FileSystem.MkdirAll(dest, pkg.FileModeDirectory)
		}

		c.AddTestFile(path, dest, mode...)
		return nil
	})
	if err != nil {
		c.T.Fatal(err)
	}
}

func (c *TestContext) AddTestDriver(src, name string) string {
	data, err := ioutil.ReadFile(src)
	if err != nil {
		c.T.Fatal(err)
	}

	dirname, err := c.FileSystem.TempDir("", "porter")
	if err != nil {
		c.T.Fatal(err)
	}

	// filename in accordance with cnab-go's command driver expectations
	filename := fmt.Sprintf("%s/cnab-%s", dirname, name)

	newfile, err := c.FileSystem.Create(filename)
	if err != nil {
		c.T.Fatal(err)
	}

	if len(data) > 0 {
		_, err := newfile.Write(data)
		if err != nil {
			c.T.Fatal(err)
		}
	}

	err = c.FileSystem.Chmod(newfile.Name(), pkg.FileModeExecutable)
	if err != nil {
		c.T.Fatal(err)
	}
	err = newfile.Close()
	if err != nil {
		c.T.Fatal(err)
	}

	path := c.Getenv("PATH")
	pathlist := []string{dirname, path}
	newpath := strings.Join(pathlist, string(os.PathListSeparator))
	c.Setenv("PATH", newpath)

	return dirname
}

// GetOutput returns all text printed to stdout.
func (c *TestContext) GetOutput() string {
	return c.capturedOut.String()
}

// GetError returns all text printed to stderr.
func (c *TestContext) GetError() string {
	return c.capturedErr.String()
}

func (c *TestContext) ClearOutputs() {
	c.capturedOut.Truncate(0)
	c.capturedErr.Truncate(0)
}

// FindRepoRoot returns the path to the porter repository where the test is currently running
func (c *TestContext) FindRepoRoot() string {
	goMod := c.findRepoFile("go.mod")
	return filepath.Dir(goMod)
}

// FindBinDir returns the path to the bin directory of the repository where the test is currently running
func (c *TestContext) FindBinDir() string {
	return c.findRepoFile("bin")
}

// Finds a file in the porter repository, does not use the mock filesystem
func (c *TestContext) findRepoFile(wantFile string) string {
	d := c.GetTestDefinitionDirectory()
	for {
		if foundFile, ok := c.hasChild(d, wantFile); ok {
			return foundFile
		}

		d = filepath.Dir(d)
		if d == "." || d == "" || d == filepath.Dir(d) {
			c.T.Fatalf("could not find %s", wantFile)
		}
	}
}

func (c *TestContext) hasChild(dir string, childName string) (string, bool) {
	children, err := ioutil.ReadDir(dir)
	if err != nil {
		c.T.Fatal(err)
	}
	for _, child := range children {
		if child.Name() == childName {
			return filepath.Join(dir, child.Name()), true
		}
	}
	return "", false
}

// CompareGoldenFile checks if the specified string matches the content of a golden test file.
// When they are different and PORTER_UPDATE_TEST_FILES is true, the file is updated to match
// the new test output.
func (c *TestContext) CompareGoldenFile(goldenFile string, got string) {
	test.CompareGoldenFile(c.T, goldenFile, got)
}
