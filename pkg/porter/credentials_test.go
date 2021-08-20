package porter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/credentials"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNoName(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

	opts := CredentialOptions{
		Silent: true,
	}
	opts.CNABFile = "/bundle.json"
	err := opts.Validate(nil, p.Porter)
	require.NoError(t, err, "Validate failed")

	err = p.GenerateCredentials(opts)
	require.NoError(t, err, "no error should have existed")

	creds, err := p.Credentials.GetCredentialSet("", "porter-hello")
	require.NoError(t, err, "expected credential to have been generated")
	var zero time.Time
	assert.True(t, zero.Before(creds.Created), "expected Credentials.Created to be set")
	assert.True(t, creds.Created.Equal(creds.Modified), "expected Credentials.Created to be initialized to Credentials.Modified")
}

func TestGenerateNameProvided(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

	opts := CredentialOptions{
		Silent: true,
	}
	opts.Namespace = "dev"
	opts.Name = "kool-kred"
	opts.Labels = []string{"env=dev"}
	opts.CNABFile = "/bundle.json"
	err := opts.Validate(nil, p.Porter)
	require.NoError(t, err, "Validate failed")

	err = p.GenerateCredentials(opts)
	require.NoError(t, err, "no error should have existed")
	creds, err := p.Credentials.GetCredentialSet(opts.Namespace, "kool-kred")
	require.NoError(t, err, "expected credential to have been generated")
	assert.Equal(t, map[string]string{"env": "dev"}, creds.Labels)
}

func TestGenerateBadNameProvided(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

	opts := CredentialOptions{
		Silent: true,
	}
	opts.Name = "this.isabadname"
	opts.CNABFile = "/bundle.json"
	err := opts.Validate(nil, p.Porter)
	require.NoError(t, err, "Validate failed")

	err = p.GenerateCredentials(opts)
	require.Error(t, err, "name is invalid, we should have had an error")
	_, err = p.Credentials.GetCredentialSet("", "this.isabadname")
	require.Error(t, err, "expected credential to not exist")
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
			wantContains: []string{"[]\n"},
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
			defer p.Teardown()

			listOpts := ListOptions{}
			listOpts.Format = tc.format
			err := p.PrintCredentials(listOpts)
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

func TestPorter_PrintCredentials(t *testing.T) {
	testcases := []CredentialsListTest{
		{
			name:         "json",
			format:       printer.FormatJson,
			wantContains: []string{"\"namespace\": \"dev\",\n    \"name\": \"kool-kreds\""},
			errorMsg:     "",
		},
		{
			name:         "yaml",
			format:       printer.FormatYaml,
			wantContains: []string{"namespace: dev\n  name: kool-kreds"},
			errorMsg:     "",
		},
		{
			name:         "table",
			format:       printer.FormatTable,
			wantContains: []string{"NAMESPACE   NAME         MODIFIED\ndev         kool-kreds   2019-06-24"},
			errorMsg:     "",
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
			defer p.Teardown()

			p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")

			listOpts := ListOptions{}
			listOpts.Namespace = "dev"
			listOpts.Format = tc.format
			err := p.PrintCredentials(listOpts)
			require.NoError(t, err)

			gotOutput := p.TestConfig.TestContext.GetOutput()
			for _, contains := range tc.wantContains {
				require.Contains(t, gotOutput, contains)
			}
		})
	}
}

// Test filtering
func TestPorter_ListCredentials(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.TestCredentials.InsertCredentialSet(credentials.NewCredentialSet("", "shared-mysql"))
	p.TestCredentials.InsertCredentialSet(credentials.NewCredentialSet("dev", "carolyn-wordpress"))
	p.TestCredentials.InsertCredentialSet(credentials.NewCredentialSet("dev", "vaughn-wordpress"))
	p.TestCredentials.InsertCredentialSet(credentials.NewCredentialSet("test", "staging-wordpress"))
	p.TestCredentials.InsertCredentialSet(credentials.NewCredentialSet("test", "iat-wordpress"))
	p.TestCredentials.InsertCredentialSet(credentials.NewCredentialSet("test", "shared-mysql"))

	t.Run("all-namespaces", func(t *testing.T) {
		opts := ListOptions{AllNamespaces: true}
		results, err := p.ListCredentials(opts)
		require.NoError(t, err)
		assert.Len(t, results, 6)
	})

	t.Run("local namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: "dev"}
		results, err := p.ListCredentials(opts)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		opts = ListOptions{Namespace: "test"}
		results, err = p.ListCredentials(opts)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("global namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: ""}
		results, err := p.ListCredentials(opts)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

type CredentialShowTest struct {
	name       string
	format     printer.Format
	wantOutput string
}

func TestShowCredential_NotFound(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	opts := CredentialShowOptions{
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatTable,
		},
		Name: "non-existent-cred",
	}

	err := p.ShowCredential(opts)
	assert.ErrorIs(t, err, storage.ErrNotFound{})
}

func TestShowCredential_Found(t *testing.T) {
	testcases := []CredentialShowTest{
		{
			name:   "json",
			format: printer.FormatJson,
			wantOutput: `{
  "schemaVersion": "1.0.0",
  "namespace": "dev",
  "name": "kool-kreds",
  "created": "2019-06-24T16:07:57.415378-05:00",
  "modified": "2019-06-24T16:07:57.415378-05:00",
  "credentials": [
    {
      "name": "kool-config",
      "source": {
        "path": "/path/to/kool-config"
      }
    },
    {
      "name": "kool-envvar",
      "source": {
        "env": "KOOL_ENV_VAR"
      }
    },
    {
      "name": "kool-cmd",
      "source": {
        "command": "echo 'kool'"
      }
    },
    {
      "name": "kool-val",
      "source": {
        "value": "kool"
      }
    }
  ]
}
`,
		},
		{
			name:   "yaml",
			format: printer.FormatYaml,
			wantOutput: `schemaVersion: 1.0.0
namespace: dev
name: kool-kreds
created: 2019-06-24T16:07:57.415378-05:00
modified: 2019-06-24T16:07:57.415378-05:00
credentials:
  - name: kool-config
    source:
      path: /path/to/kool-config
  - name: kool-envvar
    source:
      env: KOOL_ENV_VAR
  - name: kool-cmd
    source:
      command: echo 'kool'
  - name: kool-val
    source:
      value: kool

`,
		},
		{
			name:   "table",
			format: printer.FormatTable,
			wantOutput: `Name: kool-kreds
Namespace: dev
Created: 2019-06-24
Modified: 2019-06-24

--------------------------------------------------
  Name         Local Source          Source Type  
--------------------------------------------------
  kool-config  /path/to/kool-config  path         
  kool-envvar  KOOL_ENV_VAR          env          
  kool-cmd     echo 'kool'           command      
  kool-val     kool                  value        
`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Teardown()

			opts := CredentialShowOptions{
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
				Name:      "kool-kreds",
				Namespace: "dev",
			}

			p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")

			err := p.ShowCredential(opts)
			assert.NoError(t, err, "an error should not have occurred")
			gotOutput := p.TestConfig.TestContext.GetOutput()
			assert.Equal(t, tc.wantOutput, gotOutput)
		})
	}
}

func TestShowCredential_PreserveCase(t *testing.T) {
	opts := CredentialShowOptions{}
	opts.RawFormat = string(printer.FormatTable)

	err := opts.Validate([]string{"porter-hello"})
	require.NoError(t, err, "Validate failed")
	assert.Equal(t, "porter-hello", opts.Name, "Validate should preserve the credential set name case")
}

func TestCredentialsEdit(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.Setenv("SHELL", "bash")
	p.Setenv("EDITOR", "vi")
	p.Setenv(test.ExpectedCommandEnv, "bash -c vi "+filepath.Join(os.TempDir(), "porter-kool-kreds.yaml"))

	opts := CredentialEditOptions{Namespace: "dev", Name: "kool-kreds"}

	p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")
	err := p.EditCredential(opts)
	require.NoError(t, err, "no error should have existed")
}

func TestCredentialsEditEditorPathWithArgument(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	p.Setenv("SHELL", "something")
	p.Setenv("EDITOR", "C:\\Program Files\\Visual Studio Code\\code.exe --wait")
	p.Setenv(test.ExpectedCommandEnv, "something -c C:\\Program Files\\Visual Studio Code\\code.exe --wait "+filepath.Join(os.TempDir(), "porter-kool-kreds.yaml"))
	opts := CredentialEditOptions{Namespace: "dev", Name: "kool-kreds"}

	p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")
	err := p.EditCredential(opts)
	require.NoError(t, err, "no error should have existed")
}

func TestCredentialsDelete(t *testing.T) {
	testcases := []struct {
		name       string
		credName   string
		wantStderr string
	}{{
		name:     "delete",
		credName: "kool-kreds",
	}, {
		name:       "error",
		credName:   "noop-kreds",
		wantStderr: "Credential Set not found",
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Teardown()

			p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")

			opts := CredentialDeleteOptions{Namespace: "dev", Name: tc.credName}
			err := p.DeleteCredential(opts)
			require.NoError(t, err, "no error should have existed")

			_, err = p.TestCredentials.GetCredentialSet("", tc.credName)
			assert.ErrorIs(t, err, storage.ErrNotFound{})
		})
	}
}

func TestApplyOptions_Validate(t *testing.T) {
	t.Run("no file specified", func(t *testing.T) {
		tc := context.NewTestContext(t)
		opts := ApplyOptions{}
		err := opts.Validate(tc.Context, nil)
		require.EqualError(t, err, "a file argument is required")
	})

	t.Run("one file specified", func(t *testing.T) {
		tc := context.NewTestContext(t)
		tc.AddTestFileFromRoot("tests/testdata/creds/mybuns.yaml", "mybuns.yaml")
		opts := ApplyOptions{}
		err := opts.Validate(tc.Context, []string{"mybuns.yaml"})
		require.NoError(t, err)
		assert.Equal(t, "mybuns.yaml", opts.File)
	})

	t.Run("missing file specified", func(t *testing.T) {
		tc := context.NewTestContext(t)
		opts := ApplyOptions{}
		err := opts.Validate(tc.Context, []string{"mybuns.yaml"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid file argument")
	})

	t.Run("two files specified", func(t *testing.T) {
		tc := context.NewTestContext(t)
		opts := ApplyOptions{}
		err := opts.Validate(tc.Context, []string{"mybuns.yaml", "yourbuns.yaml"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "only one file argument may be specified")
	})

}
