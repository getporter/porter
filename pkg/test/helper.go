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
		if wantCmd, doAssert := os.LookupEnv(ExpectedCommandEnv); doAssert {
			gotCmd := strings.Join(os.Args[1:len(os.Args)], " ")

			if wantCmd != gotCmd {
				fmt.Printf("WANT COMMAND: %q\n", wantCmd)
				fmt.Printf("GOT COMMAND : %q\n", gotCmd)
				os.Exit(127)
			}
		}
		os.Exit(0)
	}

	// Otherwise, run the tests
	os.Exit(m.Run())
}
