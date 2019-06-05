package porter

import (
	"testing"

	cnabprovider "github.com/deislabs/porter/pkg/cnab/provider"
	printer "github.com/deislabs/porter/pkg/printer"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/stretchr/testify/assert"
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
	wantContains []string
	errorMsg     string
}

func TestCredentialsList_None(t *testing.T) {
	testcases := []CredentialsListTest{
		{
			name:         "invalid format",
			format:       "wingdings",
			wantContains: []string{},
			errorMsg:     "invalid format: wingdings",
		},
		{
			name:         "json",
			format:       printer.FormatJson,
			wantContains: []string{"[]\n"},
			errorMsg:     "",
		},
		{
			name:         "yaml",
			format:       printer.FormatYaml,
			wantContains: []string{"[]\n\n"},
			errorMsg:     "",
		},
		{
			name:         "table",
			format:       printer.FormatTable,
			wantContains: []string{"NAME   MODIFIED\n"},
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
			for _, contains := range tc.wantContains {
				require.Contains(t, gotOutput, contains)
			}
		})
	}
}

func TestCredentialsList(t *testing.T) {
	testcases := []CredentialsListTest{
		{
			name:         "json",
			format:       printer.FormatJson,
			wantContains: []string{`"Name": "kool-kreds"`},
			errorMsg:     "",
		},
		{
			name:         "yaml",
			format:       printer.FormatYaml,
			wantContains: []string{`- name: kool-kreds`},
			errorMsg:     "",
		},
		{
			name:   "table",
			format: printer.FormatTable,
			wantContains: []string{`NAME         MODIFIED
kool-kreds   now`},
			errorMsg: "",
		},
		{
			name:         "error",
			format:       printer.FormatTable,
			wantContains: []string{},
			errorMsg:     "",
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
			for _, contains := range tc.wantContains {
				require.Contains(t, gotOutput, contains)
			}
		})
	}
}

func TestCredentialsList_BadCred(t *testing.T) {
	testcases := []CredentialsListTest{
		{
			name:   "unmarshal error",
			format: printer.FormatTable,
			wantContains: []string{
				"unable to unmarshal credential set from file bad-creds.yaml: yaml: unmarshal errors",
				`NAME         MODIFIED
good-creds   now`},
			errorMsg: "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			p.CNAB = &TestCNABProvider{}

			credsDir, err := p.TestConfig.GetCredentialsDir()
			require.NoError(t, err, "no error should have existed")

			p.TestConfig.TestContext.AddTestDirectory("testdata/good-and-bad-test-creds", credsDir)

			listOpts := printer.PrintOptions{
				Format: tc.format,
			}
			err = p.ListCredentials(listOpts)
			require.NoError(t, err, "no error should have existed")

			gotOutput := p.TestConfig.TestContext.GetOutput()
			for _, contains := range tc.wantContains {
				require.Contains(t, gotOutput, contains)
			}
		})
	}
}

func TestGenerateNoCredentialDirectory(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = &TestCNABProvider{}

	opts := CredentialOptions{
		Silent: true,
	}
	opts.Name = "name"
	credDir, err := p.Config.GetCredentialsDir()
	require.NoError(t, err, "should have been able to get credentials directory path")
	credDirExists, err := p.Porter.Context.FileSystem.DirExists(credDir)
	require.NoError(t, err, "shouldn't have failed on dir exists")
	require.False(t, credDirExists, "there should not have been a credential directory for this test")
	err = p.GenerateCredentials(opts)
	assert.NoError(t, err, "credential generation should have been successful")
	credDirExists, err = p.Porter.Context.FileSystem.DirExists(credDir)
	assert.NoError(t, err, "shouldn't have gotten an error checking credential directory after generate")
	assert.True(t, credDirExists, "should have been a credential directory after the generation")
	path, err := p.Porter.Config.GetCredentialPath("name")
	assert.NoError(t, err, "couldn't get credential path")
	credFileExists, err := p.Porter.Context.FileSystem.Exists(path)
	assert.True(t, credFileExists, "expected the file %s to exist", path)
	assert.NoError(t, err, "should have been able to check if get credential path exists")
}

func TestGenerateCredentialDirectoryExists(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	p.CNAB = &TestCNABProvider{}

	opts := CredentialOptions{
		Silent: true,
	}
	opts.Name = "name"
	credDir, err := p.Config.GetCredentialsDir()
	require.NoError(t, err, "should have been able to get credentials directory path")
	err = p.Config.FileSystem.MkdirAll(credDir, 0600)
	require.NoError(t, err, "should have been able to make directory path")
	credDirExists, err := p.Porter.Context.FileSystem.DirExists(credDir)
	require.NoError(t, err, "shouldn't have failed on dir exists")
	require.True(t, credDirExists, "there should have been a credential directory for this test")
	err = p.GenerateCredentials(opts)
	assert.NoError(t, err, "credential generation should have been successful")
	credDirExists, err = p.Porter.Context.FileSystem.DirExists(credDir)
	assert.NoError(t, err, "shouldn't have gotten an error checking credential directory after generate")
	assert.True(t, credDirExists, "should have been a credential directory after the generation")
	path, err := p.Porter.Config.GetCredentialPath("name")
	assert.NoError(t, err, "couldn't get credential path")
	credFileExists, err := p.Porter.Context.FileSystem.Exists(path)
	assert.True(t, credFileExists, "expected the file %s to exist", path)
	assert.NoError(t, err, "should have been able to check if get credential path exists")
}
