package test

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const (
	MockedCommandEnv   = "MOCK_COMMAND"
	ExpectedCommandEnv = "EXPECTED_COMMAND"
)

func TestMainWithMockedCommandHandlers(m *testing.M) {
	// Fake out executing a command
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
		os.Exit(0)
	}

	// Otherwise, run the tests
	os.Exit(m.Run())
}
