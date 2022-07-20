package main

import (
	"get.porter.sh/porter/pkg/exec"
	"github.com/spf13/cobra"
)

func buildInvokeCommand(m *exec.Mixin) *cobra.Command {
	opts := exec.ExecuteOptions{}

	cmd := &cobra.Command{
		Use:   "invoke",
		Short: "Execute the invoke functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Execute(cmd.Context(), opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&opts.File, "file", "f", "", "Path to the script to execute")

	// Define a flag for --action so that its presence doesn't cause errors, but ignore it since exec doesn't need it
	var action string
	flags.StringVar(&action, "action", "", "Custom action name to invoke.")

	return cmd
}
