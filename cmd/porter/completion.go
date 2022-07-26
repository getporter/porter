package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildCompletionCommand(p *porter.Porter) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `Save the output of this command to a file and load the file into your shell.
For additional details see: https://getporter.org/install#command-completion`,
		Example:               "porter completion bash > /usr/local/etc/bash_completions.d/porter",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(p.Out)
			case "zsh":
				cmd.Root().GenZshCompletion(p.Out)
			case "fish":
				cmd.Root().GenFishCompletion(p.Out, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(p.Out)
			}
		},
	}
	cmd.Annotations = map[string]string{
		"group":    "meta",
		skipConfig: "",
	}
	return cmd
}
