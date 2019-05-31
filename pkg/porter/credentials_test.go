package porter

import (
	"testing"

	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	printer "github.com/deislabs/porter/pkg/printer"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/stretchr/testify/require"
)

type TestCNABProvider struct {
}

func (t *TestCNABProvider) LoadBundle(bundleFile string, insecure bool) (*bundle.Bundle, error) {
	b := &bundle.Bundle{
		Name: "testbundle",
		Credentials: map[string]bundle.Location{
			"name": bundle.Location{
				EnvironmentVariable: "BLAH",
			},
		},
	}
	return b, nil
}

func (t *TestCNABProvider) Install(arguments cnabprovider.InstallArguments) error {
	return nil
}
func (t *TestCNABProvider) Upgrade(arguments cnabprovider.UpgradeArguments) error {
	return nil
}
func (t *TestCNABProvider) Uninstall(arguments cnabprovider.UninstallArguments) error {
	return nil
}

func TestGenerateNoName(t *testing.T) {
	p := NewTestPorter(t)
	p.CNAB = &TestCNABProvider{}

	opts := CredentialOptions{
		Silent: true,
	}
	err := p.GenerateCredentials(opts)
	require.NoError(t, err, "no error should have existed")
	path, err := p.Porter.Config.GetCredentialPath("testbundle")
	require.NoError(t, err, "couldn't get credential path")
	_, err = p.Porter.Context.FileSystem.Stat(path)
	require.NoError(t, err, "expected the file %s to exist", path)
}

func TestGenerateNameProvided(t *testing.T) {
	p := NewTestPorter(t)
	p.CNAB = &TestCNABProvider{}

	opts := CredentialOptions{
		Silent: true,
	}
	opts.Name = "kool-kred"

	err := p.GenerateCredentials(opts)
	require.NoError(t, err, "no error should have existed")
	path, err := p.Porter.Config.GetCredentialPath("kool-kred")
	require.NoError(t, err, "couldn't get credential path")
	_, err = p.Porter.Context.FileSystem.Stat(path)
	require.NoError(t, err, "expected the file %s to exist", path)
}

func TestGenerateBadNameProvided(t *testing.T) {
	p := NewTestPorter(t)
	p.CNAB = &TestCNABProvider{}

	opts := CredentialOptions{
		Silent: true,
	}
	opts.Name = "this.isabadname"

	err := p.GenerateCredentials(opts)
	require.Error(t, err, "name is invalid, we should have had an error")
	path, err := p.Porter.Config.GetCredentialPath("this.isabadname")
	require.NoError(t, err, "couldn't get credential path")
	_, err = p.Porter.Context.FileSystem.Stat(path)
	require.Error(t, err, "expected the file %s to not exist", path)
}

type CredentialsListTest struct {
	name         string
	format       printer.Format
	wantContains string
	errorMsg     string
}

func TestCredentialsList_None(t *testing.T) {
	testcases := []CredentialsListTest{
		{
			name:         "invalid format",
			format:       "wingdings",
			wantContains: "",
			errorMsg:     "invalid format: wingdings",
		},
		{
			name:         "json",
			format:       printer.FormatJson,
			wantContains: "[]\n",
			errorMsg:     "",
		},
		{
			name:         "yaml",
			format:       printer.FormatYaml,
			wantContains: "[]\n\n",
			errorMsg:     "",
		},
		{
			name:         "table",
			format:       printer.FormatTable,
			wantContains: "NAME   MODIFIED\n",
			errorMsg:     "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			p.CNAB = &TestCNABProvider{}

			listOpts := printer.PrintOptions{
				Format: tc.format,
			}
			err := p.ListCredentials(listOpts)
			if tc.errorMsg != "" {
				require.Equal(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err, "no error should have existed")
			}

			gotOutput := p.TestConfig.TestContext.GetOutput()
			require.Equal(t, tc.wantContains, gotOutput)
		})
	}
}

func TestCredentialsList(t *testing.T) {
	testcases := []CredentialsListTest{
		{
			name:         "json",
			format:       printer.FormatJson,
			wantContains: `"Name": "kool-kreds"`,
			errorMsg:     "",
		},
		{
			name:         "yaml",
			format:       printer.FormatYaml,
			wantContains: `- name: kool-kreds`,
			errorMsg:     "",
		},
		{
			name:   "table",
			format: printer.FormatTable,
			wantContains: `NAME         MODIFIED
kool-kreds   now`,
			errorMsg: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			p.CNAB = &TestCNABProvider{}

			credsDir, err := p.TestConfig.GetCredentialsDir()
			require.NoError(t, err, "no error should have existed")

			p.TestConfig.TestContext.AddTestDirectory("testdata/test-creds", credsDir)

			listOpts := printer.PrintOptions{
				Format: tc.format,
			}
			err = p.ListCredentials(listOpts)
			require.NoError(t, err, "no error should have existed")

			gotOutput := p.TestConfig.TestContext.GetOutput()
			// TODO: change to require.Equal, verify modified, perhaps w/ regex?
			require.Contains(t, gotOutput, tc.wantContains)
		})
	}
}
