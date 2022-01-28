package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	MockedCommandEnv           = "MOCK_COMMAND"
	ExpectedCommandEnv         = "EXPECTED_COMMAND"
	ExpectedCommandExitCodeEnv = "EXPECTED_COMMAND_EXIT_CODE"
	ExpectedCommandErrorEnv    = "EXPECTED_COMMAND_ERROR"
)

func TestMainWithMockedCommandHandlers(m *testing.M) {

	// Fake out executing a command
	// It's okay to use os.LookupEnv here because it's running in it's own process, and won't impact running tests in parallel.
	if _, mockCommand := os.LookupEnv(MockedCommandEnv); mockCommand {
		if expectedCmdEnv, doAssert := os.LookupEnv(ExpectedCommandEnv); doAssert {
			gotCmd := strings.Join(os.Args[1:len(os.Args)], " ")

			// There may be multiple expected commands, separated by a newline character
			wantCmds := strings.Split(expectedCmdEnv, "\n")

			commandNotFound := true
			for _, wantCmd := range wantCmds {
				if wantCmd == gotCmd {
					commandNotFound = false
				}
			}

			if commandNotFound {
				fmt.Printf("WANT COMMANDS : %q\n", wantCmds)
				fmt.Printf("GOT COMMAND : %q\n", gotCmd)
				os.Exit(127)
			}
		}

		if wantError, ok := os.LookupEnv(ExpectedCommandErrorEnv); ok {
			fmt.Fprintln(os.Stderr, wantError)
		}

		exitCode := 0
		if wantCode, ok := os.LookupEnv(ExpectedCommandExitCodeEnv); ok {
			exitCode, _ = strconv.Atoi(wantCode)
		}
		os.Exit(exitCode)
	}

	// Otherwise, run the tests
	os.Exit(m.Run())
}

// CompareGoldenFile checks if the specified string matches the content of a golden test file.
// When they are different and PORTER_UPDATE_TEST_FILES is true, the file is updated to match
// the new test output.
func CompareGoldenFile(t *testing.T, goldenFile string, got string) {
	if os.Getenv("PORTER_UPDATE_TEST_FILES") == "true" {
		os.MkdirAll(filepath.Dir(goldenFile), 0700)
		t.Logf("Updated test file %s to match latest test output", goldenFile)
		require.NoError(t, ioutil.WriteFile(goldenFile, []byte(got), 0600), "could not update golden file %s", goldenFile)
	} else {
		wantSchema, err := ioutil.ReadFile(goldenFile)
		require.NoError(t, err)
		assert.Equal(t, string(wantSchema), got, "The test output doesn't match the expected output in %s. If this was intentional, run mage updateTestfiles to fix the tests.", goldenFile)
	}
}
