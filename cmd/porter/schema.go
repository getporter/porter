package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildSchemaCommand(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Print the JSON schema for the Porter manifest",
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.PrintManifestSchema()
		},
	}
	cmd.Annotations = map[string]string{
		"group": "meta",
	}
	return cmd
}
