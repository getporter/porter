package main

import (
	"io"

	"github.com/spf13/cobra"
)

func newPushCmd(out io.Writer) *cobra.Command {
	const usage = `Pushes a CNAB bundle to a repository.`

	cmd := &cobra.Command{
		Hidden: true,
		Use:    "push NAME",
		Short:  "push a CNAB bundle to a repository",
		Long:   usage,
		Args:   cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return ErrUnderConstruction
		},
	}

	return cmd
}
