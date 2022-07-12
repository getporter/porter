package porter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateNoName(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()
	ctx := context.Background()

	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

	opts := CredentialOptions{
		Silent: true,
	}
	opts.CNABFile = "/bundle.json"
	err := opts.Validate(ctx, nil, p.Porter)
	require.NoError(t, err, "Validate failed")

	err = p.GenerateCredentials(ctx, opts)
	require.NoError(t, err, "no error should have existed")

	creds, err := p.Credentials.GetCredentialSet(ctx, "", "porter-hello")
	require.NoError(t, err, "expected credential to have been generated")
	var zero time.Time
	assert.True(t, zero.Before(creds.Status.Created), "expected Credentials.Created to be set")
	assert.True(t, creds.Status.Created.Equal(creds.Status.Modified), "expected Credentials.Created to be initialized to Credentials.Modified")
}

func TestGenerateNameProvided(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()
	ctx := context.Background()

	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

	opts := CredentialOptions{
		Silent: true,
	}
	opts.Namespace = "dev"
	opts.Name = "kool-kred"
	opts.Labels = []string{"env=dev"}
	opts.CNABFile = "/bundle.json"
	err := opts.Validate(ctx, nil, p.Porter)
	require.NoError(t, err, "Validate failed")

	err = p.GenerateCredentials(ctx, opts)
	require.NoError(t, err, "no error should have existed")
	creds, err := p.Credentials.GetCredentialSet(ctx, opts.Namespace, "kool-kred")
	require.NoError(t, err, "expected credential to have been generated")
	assert.Equal(t, map[string]string{"env": "dev"}, creds.Labels)
}

func TestGenerateBadNameProvided(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()
	ctx := context.Background()

	p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

	opts := CredentialOptions{
		Silent: true,
	}
	opts.Name = "this.isabadname"
	opts.CNABFile = "/bundle.json"
	err := opts.Validate(ctx, nil, p.Porter)
	require.NoError(t, err, "Validate failed")

	err = p.GenerateCredentials(ctx, opts)
	require.Error(t, err, "name is invalid, we should have had an error")
	_, err = p.Credentials.GetCredentialSet(ctx, "", "this.isabadname")
	require.Error(t, err, "expected credential to not exist")
}

type CredentialsListTest struct {
	name       string
	format     printer.Format
	wantOutput string
	errorMsg   string
}

func TestCredentialsList_None(t *testing.T) {
	ctx := context.Background()

	testcases := []CredentialsListTest{
		{
			name:     "invalid format",
			format:   "wingdings",
			errorMsg: "invalid format: wingdings",
		},
		{
			name:       "json",
			format:     printer.FormatJson,
			wantOutput: "testdata/credentials/list-output.json",
			errorMsg:   "",
		},
		{
			name:       "yaml",
			format:     printer.FormatYaml,
			wantOutput: "testdata/credentials/list-output.yaml",
			errorMsg:   "",
		},
		{
			name:       "plaintext",
			format:     printer.FormatPlaintext,
			wantOutput: "testdata/credentials/list-output.txt",
			errorMsg:   "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			listOpts := ListOptions{}
			listOpts.Format = tc.format
			err := p.PrintCredentials(ctx, listOpts)
			if tc.errorMsg != "" {
				require.Equal(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err, "no error should have existed")
				gotOutput := p.TestConfig.TestContext.GetOutput()
				test.CompareGoldenFile(t, tc.wantOutput, gotOutput)
			}
		})
	}
}

func TestPorter_PrintCredentials(t *testing.T) {
	ctx := context.Background()

	testcases := []CredentialsListTest{
		{
			name:       "json",
			format:     printer.FormatJson,
			wantOutput: "testdata/credentials/show-output.json",
		},
		{
			name:       "yaml",
			format:     printer.FormatYaml,
			wantOutput: "testdata/credentials/show-output.yaml",
		},
		{
			name:       "plaintext",
			format:     printer.FormatPlaintext,
			wantOutput: "testdata/credentials/show-output.txt",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")

			listOpts := ListOptions{}
			listOpts.Namespace = "dev"
			listOpts.Format = tc.format
			err := p.PrintCredentials(ctx, listOpts)
			require.NoError(t, err)

			gotOutput := p.TestConfig.TestContext.GetOutput()
			test.CompareGoldenFile(t, tc.wantOutput, gotOutput)
		})
	}
}

// Test filtering
func TestPorter_ListCredentials(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	ctx := context.Background()
	p.TestCredentials.InsertCredentialSet(ctx, storage.NewCredentialSet("", "shared-mysql"))
	p.TestCredentials.InsertCredentialSet(ctx, storage.NewCredentialSet("dev", "carolyn-wordpress"))
	p.TestCredentials.InsertCredentialSet(ctx, storage.NewCredentialSet("dev", "vaughn-wordpress"))
	p.TestCredentials.InsertCredentialSet(ctx, storage.NewCredentialSet("test", "staging-wordpress"))
	p.TestCredentials.InsertCredentialSet(ctx, storage.NewCredentialSet("test", "iat-wordpress"))
	p.TestCredentials.InsertCredentialSet(ctx, storage.NewCredentialSet("test", "shared-mysql"))

	t.Run("all-namespaces", func(t *testing.T) {
		opts := ListOptions{AllNamespaces: true}
		results, err := p.ListCredentials(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 6)
	})

	t.Run("local namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: "dev"}
		results, err := p.ListCredentials(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		opts = ListOptions{Namespace: "test"}
		results, err = p.ListCredentials(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("global namespace", func(t *testing.T) {
		opts := ListOptions{Namespace: ""}
		results, err := p.ListCredentials(ctx, opts)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestShowCredential_NotFound(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	opts := CredentialShowOptions{
		PrintOptions: printer.PrintOptions{
			Format: printer.FormatPlaintext,
		},
		Name: "non-existent-cred",
	}

	err := p.ShowCredential(context.Background(), opts)
	assert.ErrorIs(t, err, storage.ErrNotFound{})
}

func TestShowCredential_Found(t *testing.T) {
	type CredentialShowTest struct {
		name               string
		format             printer.Format
		expectedOutputFile string
	}

	testcases := []CredentialShowTest{
		{
			name:               "json",
			format:             printer.FormatJson,
			expectedOutputFile: "testdata/credentials/kool-kreds.json",
		},
		{
			name:               "yaml",
			format:             printer.FormatYaml,
			expectedOutputFile: "testdata/credentials/kool-kreds.yaml",
		},
		{
			name:               "table",
			format:             printer.FormatPlaintext,
			expectedOutputFile: "testdata/credentials/kool-kreds.txt",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			opts := CredentialShowOptions{
				PrintOptions: printer.PrintOptions{
					Format: tc.format,
				},
				Name:      "kool-kreds",
				Namespace: "dev",
			}

			p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")

			err := p.ShowCredential(context.Background(), opts)
			require.NoError(t, err, "an error should not have occurred")
			gotOutput := p.TestConfig.TestContext.GetOutput()
			test.CompareGoldenFile(t, tc.expectedOutputFile, gotOutput)
		})
	}
}

func TestShowCredential_PreserveCase(t *testing.T) {
	opts := CredentialShowOptions{}
	opts.RawFormat = string(printer.FormatPlaintext)

	err := opts.Validate([]string{"porter-hello"})
	require.NoError(t, err, "Validate failed")
	assert.Equal(t, "porter-hello", opts.Name, "Validate should preserve the credential set name case")
}

func TestCredentialsEdit(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.Setenv("SHELL", "bash")
	p.Setenv("EDITOR", "vi")
	p.Setenv(test.ExpectedCommandEnv, "bash -c vi "+filepath.Join(os.TempDir(), "porter-kool-kreds.yaml"))

	opts := CredentialEditOptions{Namespace: "dev", Name: "kool-kreds"}

	p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")
	err := p.EditCredential(context.Background(), opts)
	require.NoError(t, err, "no error should have existed")
}

func TestCredentialsEditEditorPathWithArgument(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.Setenv("SHELL", "something")
	p.Setenv("EDITOR", "C:\\Program Files\\Visual Studio Code\\code.exe --wait")
	p.Setenv(test.ExpectedCommandEnv, "something -c C:\\Program Files\\Visual Studio Code\\code.exe --wait "+filepath.Join(os.TempDir(), "porter-kool-kreds.yaml"))
	opts := CredentialEditOptions{Namespace: "dev", Name: "kool-kreds"}

	p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")
	err := p.EditCredential(context.Background(), opts)
	require.NoError(t, err, "no error should have existed")
}

func TestCredentialsDelete(t *testing.T) {
	ctx := context.Background()

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
			defer p.Close()

			p.TestCredentials.AddTestCredentialsDirectory("testdata/test-creds")

			opts := CredentialDeleteOptions{Namespace: "dev", Name: tc.credName}
			err := p.DeleteCredential(ctx, opts)
			require.NoError(t, err, "no error should have existed")

			_, err = p.TestCredentials.GetCredentialSet(ctx, "", tc.credName)
			assert.ErrorIs(t, err, storage.ErrNotFound{})
		})
	}
}

func TestApplyOptions_Validate(t *testing.T) {
	t.Run("no file specified", func(t *testing.T) {
		tc := portercontext.NewTestContext(t)
		opts := ApplyOptions{}
		err := opts.Validate(tc.Context, nil)
		require.EqualError(t, err, "a file argument is required")
	})

	t.Run("one file specified", func(t *testing.T) {
		tc := portercontext.NewTestContext(t)
		tc.AddTestFileFromRoot("tests/testdata/creds/mybuns.yaml", "mybuns.yaml")
		opts := ApplyOptions{}
		err := opts.Validate(tc.Context, []string{"mybuns.yaml"})
		require.NoError(t, err)
		assert.Equal(t, "mybuns.yaml", opts.File)
	})

	t.Run("missing file specified", func(t *testing.T) {
		tc := portercontext.NewTestContext(t)
		opts := ApplyOptions{}
		err := opts.Validate(tc.Context, []string{"mybuns.yaml"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid file argument")
	})

	t.Run("two files specified", func(t *testing.T) {
		tc := portercontext.NewTestContext(t)
		opts := ApplyOptions{}
		err := opts.Validate(tc.Context, []string{"mybuns.yaml", "yourbuns.yaml"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "only one file argument may be specified")
	})

}

func TestCredentialsCreateOptions_Validate(t *testing.T) {
	testcases := []struct {
		name       string
		args       []string
		outputType string
		wantErr    string
	}{
		{
			name:       "no fileName defined",
			args:       []string{},
			outputType: "",
			wantErr:    "",
		},
		{
			name:       "two positional arguments",
			args:       []string{"credential-set1", "credential-set2"},
			outputType: "",
			wantErr:    "only one positional argument may be specified",
		},
		{
			name:       "no file format defined from file extension or output flag",
			args:       []string{"credential-set"},
			outputType: "",
			wantErr:    "could not detect the file format from the file extension (.txt). Specify the format with --output",
		},
		{
			name:       "different file format",
			args:       []string{"credential-set.json"},
			outputType: "yaml",
			wantErr:    "",
		},
		{
			name:       "format from output flag",
			args:       []string{"creds"},
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "format from file extension",
			args:       []string{"credential-set.yml"},
			outputType: "",
			wantErr:    "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := CredentialCreateOptions{OutputType: tc.outputType}
			err := opts.Validate(tc.args)
			if tc.wantErr == "" {
				require.NoError(t, err, "no error should have existed")
				return
			}
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestCredentialsCreate(t *testing.T) {
	testcases := []struct {
		name       string
		fileName   string
		outputType string
		wantErr    string
	}{
		{
			name:       "valid input: no input defined, will output yaml format to stdout",
			fileName:   "",
			outputType: "",
			wantErr:    "",
		},
		{
			name:       "valid input: output to stdout with format json",
			fileName:   "",
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "valid input: file format from fileName",
			fileName:   "fileName.json",
			outputType: "",
			wantErr:    "",
		},
		{
			name:       "valid input: file format from outputType",
			fileName:   "fileName",
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "valid input: different file format from fileName and outputType",
			fileName:   "fileName.yaml",
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "valid input: same file format in fileName and outputType",
			fileName:   "fileName.json",
			outputType: "json",
			wantErr:    "",
		},
		{
			name:       "invalid input: invalid file format from fileName",
			fileName:   "fileName.txt",
			outputType: "",
			wantErr:    fmt.Sprintf("unsupported format %s. Supported formats are: yaml and json", "txt"),
		},
		{
			name:       "invalid input: invalid file format from outputType",
			fileName:   "fileName",
			outputType: "txt",
			wantErr:    fmt.Sprintf("unsupported format %s. Supported formats are: yaml and json", "txt"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			p := NewTestPorter(t)
			defer p.Close()

			opts := CredentialCreateOptions{FileName: tc.fileName, OutputType: tc.outputType}
			err := p.CreateCredential(ctx, opts)
			if tc.wantErr == "" {
				require.NoError(t, err, "no error should have existed")
				return
			}
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
