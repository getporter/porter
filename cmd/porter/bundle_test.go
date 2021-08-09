package main

import (
	"strings"
	"testing"

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
