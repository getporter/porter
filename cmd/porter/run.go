package main

import (
	"fmt"
	"os"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildRunCommand(p *porter.Porter) *cobra.Command {
	var opts struct {
		file      string
		rawAction string
		action    config.Action
	}
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute runtime bundle instructions",
		Long:  "Execute the runtime bundle instructions contained in a porter configuration file",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.rawAction == "" {
				opts.rawAction = os.Getenv("CNAB_ACTION")
				if p.Debug {
					fmt.Fprintf(p.Out, "DEBUG: defaulting action to CNAB_ACTION (%s)\n", opts.rawAction)
				}
			}

			var err error
			opts.action, err = config.ParseAction(opts.rawAction)
			if err != nil {
				return err
			}

			if exists, _ := p.FileSystem.Exists(opts.file); !exists {
				return fmt.Errorf("invalid --file: the specified porter configuration file %q doesn't exist", opts.file)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.Run(opts.file, opts.action)
		},
	}

	cmd.Flags().StringVarP(&opts.file, "file", "f", "porter.yaml", "The porter configuration file (Defaults to porter.yaml)")
	cmd.Flags().StringVar(&opts.rawAction, "action", "", "The bundle action to execute (Defaults to CNAB_ACTION)")

	return cmd
}
