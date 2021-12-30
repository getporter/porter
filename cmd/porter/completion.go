package main

import (
	"os"

	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildCompletionCommand(p *porter.Porter) *cobra.Command {

	cmd := &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate completion script",
		Long:                  "Capture the output of the completion command to a file for your shell environment.",
		Example:               "porter completion bash > /usr/local/etc/bash_completions.d/porter",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
	cmd.Annotations = map[string]string{
		"group": "meta",
	}
	return cmd
}
