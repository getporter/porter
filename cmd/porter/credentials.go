package main

import (
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
)

func buildCredentialsCommands(p *porter.Porter) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "credentials",
		Aliases:     []string{"credential", "cred", "creds"},
		Annotations: map[string]string{"group": "resource"},
		Short:       "Credentials commands",
	}

	cmd.AddCommand(buildCredentialsEditCommand(p))
	cmd.AddCommand(buildCredentialsGenerateCommand(p))
	cmd.AddCommand(buildCredentialsListCommand(p))
	cmd.AddCommand(buildCredentialsDeleteCommand(p))
	cmd.AddCommand(buildCredentialsShowCommand(p))

	return cmd
}

func buildCredentialsEditCommand(p *porter.Porter) *cobra.Command {
	opts := porter.CredentialEditOptions{}

	cmd := &cobra.Command{
		Use:     "edit",
		Short:   "Edit Credential",
		Long:    `Edit a named credential set.`,
		Example: `  porter credentials edit github --namespace dev`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.EditCredential(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the credential set is defined. Defaults to the global namespace.")

	return cmd
}

func buildCredentialsGenerateCommand(p *porter.Porter) *cobra.Command {
	opts := porter.CredentialOptions{}
	cmd := &cobra.Command{
		Use:   "generate [NAME]",
		Short: "Generate Credential Set",
		Long: `Generate a named set of credentials.

The first argument is the name of credential set you wish to generate. If not
provided, this will default to the bundle name. By default, Porter will
generate a credential set for the bundle in the current directory. You may also
specify a bundle with --file.

Bundles define 1 or more credential(s) that are required to interact with a
bundle. The bundle definition defines where the credential should be delivered
to the bundle, i.e. at /root/.kube. A credential set, on the other hand,
represents the source data that you wish to use when interacting with the
bundle. These will typically be environment variables or files on your local
file system.

When you wish to install, upgrade or delete a bundle, Porter will use the
credential set to determine where to read the necessary information from and
will then provide it to the bundle in the correct location. `,
		Example: `  porter credential generate
  porter credential generate kubecred --reference getporter/porter-hello:v0.1.0 --namespace test
  porter credential generate kubecred --reference localhost:5000/getporter/porter-hello:v0.1.0 --insecure-registry --force
  porter credential generate kubecred --file myapp/porter.yaml
  porter credential generate kubecred --cnab-file myapp/bundle.json
`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args, p.Context)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.GenerateCredentials(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the credential set is defined. Defaults to the global namespace.")
	f.StringVarP(&opts.File, "file", "f", "",
		"Path to the porter manifest file. Defaults to the bundle in the current directory.")
	f.StringVar(&opts.CNABFile, "cnab-file", "",
		"Path to the CNAB bundle.json file.")
	addBundlePullFlags(f, &opts.BundlePullOptions)

	return cmd
}

func buildCredentialsListCommand(p *porter.Porter) *cobra.Command {
	opts := porter.ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List credentials",
		Long:    `List named sets of credentials defined by the user.`,
		Example: `  porter credentials list
  porter credentials list --namespace prod
  porter credentials list --namespace "*"`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.ParseFormat()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ListCredentials(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the credential set is defined. Defaults to the global namespace. Use * to list across all namespaces.")
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return cmd
}

func buildCredentialsDeleteCommand(p *porter.Porter) *cobra.Command {
	opts := porter.CredentialDeleteOptions{}

	cmd := &cobra.Command{
		Use:     "delete NAME",
		Short:   "Delete a Credential",
		Long:    `Delete a named credential set.`,
		Example: `  porter credentials delete github --namespace dev`,
		PreRunE: func(_ *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return p.DeleteCredential(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the credential set is defined. Defaults to the global namespace.")

	return cmd
}

func buildCredentialsShowCommand(p *porter.Porter) *cobra.Command {
	opts := porter.CredentialShowOptions{}

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show a Credential",
		Long:  `Show a particular credential set, including all named credentials and their corresponding mappings.`,
		Example: `  porter credential show github --namespace dev
  porter credential show prodcluster --output json`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return opts.Validate(args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return p.ShowCredential(opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.Namespace, "namespace", "n", "",
		"Namespace in which the credential set is defined. Defaults to the global namespace.")
	f.StringVarP(&opts.RawFormat, "output", "o", "table",
		"Specify an output format.  Allowed values: table, json, yaml")

	return cmd
}
