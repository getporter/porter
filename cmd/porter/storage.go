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
		Use:   "migrate --old-home OLD_PORTER_HOME [--old-account STORAGE_NAME] [--namespace NAMESPACE]",
		Short: "Migrate data from v0.38 to v1",
		Long: `Migrate data from Porter v0.38 into a v1 installation of Porter.

See https://getporter.org/storage-migrate for a full description of the migration process. Below is a summary:

Before running this command, you should have:

1. Installed Porter v1, see https://getporter.org/install for instructions.
2. Renamed the old PORTER_HOME directory, for example: mv ~/.porter ~/.porterv0.
3. Created a new PORTER_HOME directory for the new version of Porter, for example: mkdir ~/.porter.
4. Configured a default storage account for the new version of Porter. The data from the old account will be migrated into the default storage account.

This upgrades the data to the current storage schema, and does not change the data stored in the old account.`,
		Example: `  porter storage migrate --old-home ~/.porterv0
  porter storage migrate --old-account my-azure --old-home ~/.porterv0
  porter storage migrate --namespace new-namespace --old-home ~/.porterv0
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.MigrateStorage(cmd.Context(), opts)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&opts.OldHome, "old-home", "",
		"Path to the old PORTER_HOME directory where the previous version of Porter is installed")
	flags.StringVar(&opts.OldStorageAccount, "old-account", "",
		"Name of the storage account in the old Porter configuration file containing the data that should be migrated. If unspecified, the default storage account is used.")
	flags.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Destination namespace where the migrated data should be saved. By default, Porter migrates your data into the current namespace as configured by environment variables and your config file, otherwise the global namespace is used.")
	return cmd
}

func buildStorageFixPermissionsCommand(p *porter.Porter) *cobra.Command {
	return &cobra.Command{
		Use:   "fix-permissions",
		Short: "Fix the permissions on your PORTER_HOME directory",
		Long:  `This will reset the permissions on your PORTER_HOME directory to the least permissions required, where only the current user has permissions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.FixPermissions(cmd.Context())
		},
	}
}
