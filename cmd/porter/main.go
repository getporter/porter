package main

import (
	"os"

	"github.com/deislabs/porter/pkg/porter"

	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCommand() *cobra.Command {
	p := porter.New()
	cmd := &cobra.Command{
		Use:  "porter",
		Long: "I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			p.Out = cmd.OutOrStdout()
		},
	}

	cmd.AddCommand(buildVersionCommand(p))
	cmd.AddCommand(buildInitCommand(p))
	cmd.AddCommand(buildRunCommand(p))

	return cmd
}
