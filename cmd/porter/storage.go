package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
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
	cmd.AddCommand(buildStorageFixPermissionsCommand(p))

	return &cmd
}

func buildStorageMigrateCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate active storage account",
		Long: `Migrate the data in the active storage account to the schema used by this version of Porter.

Always back up Porter's data before performing a migration. Instructions for backing up are at https://getporter.org/storage-migrate.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.MigrateStorage(cmd.Context())
		},
	}
}

func buildStorageFixPermissionsCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:   "fix-permissions",
		Short: "Fix the permissions on your PORTER_HOME directory",
		Long:  `This will reset the permissions on your PORTER_HOME directory to the least permissions required, where only the current user has permissions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.FixPermissions()
		},
	}
}
