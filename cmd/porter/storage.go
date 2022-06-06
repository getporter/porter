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
	var opts porter.MigrateStorageOptions
	cmd := &cobra.Command{
		Use:   "migrate --src OLD_ACCOUNT --dest NEW_ACCOUNT",
		Short: "Migrate data from an older version of Porter",
		Long: `Copies data from a source storage account defined in Porter's config file into a destination storage account. 

This upgrades the data to the current storage schema, and does not change the data stored in the source account.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.MigrateStorage(cmd.Context(), opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.Source, "src", "s", "",
		"Name of the source storage account defined in your Porter config file")
	flags.StringVarP(&opts.Destination, "dest", "d", "",
		"Name of the destination storage account defined in your Porter config file")
	return cmd
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
