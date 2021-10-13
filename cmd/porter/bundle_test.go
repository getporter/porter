package main

import (
	"os"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/porter"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateInstallCommand(t *testing.T) {
	testcases := []struct {
		name      string
		args      string
		wantError string
	}{
		{"no args", "install -r getporter/porter-hello:v0.1.1", ""},
		{"invalid param", "install --param A:B -r getporter/porter-hello:v0.1.1", "invalid parameter (A:B), must be in name=value format"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := buildRootCommand()
			osargs := strings.Split(tc.args, " ")
			cmd, args, err := p.Find(osargs)
			require.NoError(t, err)

			err = cmd.ParseFlags(args)
			require.NoError(t, err)

			err = cmd.PreRunE(cmd, cmd.Flags().Args())
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestValidateUninstallCommand(t *testing.T) {
	testcases := []struct {
		name      string
		args      string
		wantError string
	}{
		{"no args", "uninstall mybuns", ""},
		{"invalid param", "uninstall mybuns --param A:B", "invalid parameter (A:B), must be in name=value format"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := buildRootCommand()
			osargs := strings.Split(tc.args, " ")
			cmd, args, err := p.Find(osargs)
			require.NoError(t, err)

			err = cmd.ParseFlags(args)
			require.NoError(t, err)

			err = cmd.PreRunE(cmd, cmd.Flags().Args())
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestValidateInvokeCommand(t *testing.T) {
	testcases := []struct {
		name      string
		args      string
		wantError string
	}{
		{"no action", "invoke mybuns", "--action is required"},
		{"action specified", "invoke mybuns --action status", ""},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := buildRootCommand()
			osargs := strings.Split(tc.args, " ")
			cmd, args, err := p.Find(osargs)
			require.NoError(t, err)

			err = cmd.ParseFlags(args)
			require.NoError(t, err)

			err = cmd.PreRunE(cmd, cmd.Flags().Args())
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestValidateInstallationListCommand(t *testing.T) {
	testcases := []struct {
		name      string
		args      string
		wantError string
	}{
		{"no args", "installation list", ""},
		{"output json", "installation list -o json", ""},
		{"invalid format", "installation list -o wingdings", "invalid format: wingdings"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := buildRootCommand()
			osargs := strings.Split(tc.args, " ")
			cmd, args, err := p.Find(osargs)
			require.NoError(t, err)

			err = cmd.ParseFlags(args)
			require.NoError(t, err)

			err = cmd.PreRunE(cmd, cmd.Flags().Args())
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}

func TestBuildValidate_Driver(t *testing.T) {
	// Do not run in parallel
	os.Setenv("PORTER_EXPERIMENTAL", experimental.BuildDrivers)
	defer os.Unsetenv("PORTER_EXPERIMENTAL")

	testcases := []struct {
		name         string
		args         string
		configDriver string // the driver set in the config
		wantDriver   string
		wantError    string
	}{
		{name: "no flag", wantDriver: "docker"},
		{name: "invalid flag", args: "--driver=missing-driver", wantError: "invalid --driver value missing-driver"},
		{name: "valid flag", args: "--driver=buildkit", wantDriver: "buildkit"},
		{name: "invalid config", args: "", configDriver: "invalid-driver", wantError: "invalid --driver value invalid-driver"},
		{name: "valid config", args: "", configDriver: "buildkit", wantDriver: "buildkit"}, // passing an empty flag to trigger defaulting it to the config value
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.configDriver != "" {
				os.Setenv("PORTER_BUILD_DRIVER", tc.configDriver)
				defer os.Unsetenv("PORTER_BUILD_DRIVER")
			}

			p := porter.NewTestPorter(t)
			defer p.Teardown()

			rootCmd := buildRootCommandFrom(p.Porter)

			fullArgs := []string{"build", tc.args}
			rootCmd.SetArgs(fullArgs)
			buildCmd, _, _ := rootCmd.Find(fullArgs)
			buildCmd.RunE = func(cmd *cobra.Command, args []string) error {
				// noop
				return nil
			}

			err := rootCmd.Execute()
			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantDriver, p.Data.BuildDriver)
			} else {
				require.EqualError(t, err, tc.wantError)
			}
		})
	}
}
