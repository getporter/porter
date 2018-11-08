package main

import (
	"fmt"
	"os"

	"github.com/deislabs/porter/pkg/mixin/exec"
	"github.com/spf13/cobra"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				if err != nil {
					fmt.Printf("panic, err: %s\n", err)
					os.Exit(1)
				}
			}
		}
	}()
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
}

func buildRootCommand() *cobra.Command {
	e := &exec.Exec{}
	cmd := &cobra.Command{
		Use:  "exec",
		Long: "exec is a porter 👩🏽‍✈️ mixin that you can you can use to execute arbitrary commands",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			e.Out = cmd.OutOrStdout()
			e.Err = cmd.OutOrStderr()
		},
	}

	cmd.AddCommand(buildVersionCommand(e))
	cmd.AddCommand(buildInstallCommand(e))
	return cmd
}
