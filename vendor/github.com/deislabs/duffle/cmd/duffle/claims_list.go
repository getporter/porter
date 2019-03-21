package main

import (
	"io"

	"github.com/spf13/cobra"
)

func newClaimListCmd(out io.Writer) *cobra.Command {
	list := listCmd{out: out}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "list available claims",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := &listCmd{out: out, short: list.short}
			return l.run()
		},
	}

	f := cmd.Flags()
	f.BoolVarP(&list.short, "short", "s", false, "output shorter listing format")

	return cmd
}
