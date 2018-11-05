package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "porter",
		Long: "I am porter 👩🏽‍✈️, the friendly neighborhood CNAB authoring tool",
	}

	cmd.AddCommand(buildVersionCommand())

	return cmd
}
