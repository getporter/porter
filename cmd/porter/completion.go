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
For additional details see: https://porter.sh/install#command-completion`,
		Example:               "porter completion bash > /usr/local/etc/bash_completions.d/porter",
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(p.Out)
			case "zsh":
				return cmd.Root().GenZshCompletion(p.Out)
			case "fish":
				return cmd.Root().GenFishCompletion(p.Out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(p.Out)
			}
			return nil
		},
	}
	cmd.Annotations = map[string]string{
		"group":    "meta",
		skipConfig: "",
	}
	return cmd
}
