package main

import (
	"github.com/spf13/cobra"

	"get.porter.sh/porter/pkg/porter"
)

func buildStorageCommand(p *porter.Porter) *cobra.Command {
	cmd := cobra.Command{
		Use:   "storage",
		Short: "Manage data stored by Porter",
		Long: `Manage the data stored by Porter, such as credentials and installation data.
`,
		Annotations: map[string]string{
			"group": "resource",
		},
	}

	cmd.AddCommand(buildStorageMigrateCommand(p))

	return &cmd
}

func buildStorageMigrateCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate active storage account",
		Long: `Migrate the data in the active storage account to the schema used by this version of Porter.

Always back up Porter's data before performing a migration. Instructions for backing up are at https://porter.sh/storage-migrate.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.MigrateStorage()
		},
	}
}
