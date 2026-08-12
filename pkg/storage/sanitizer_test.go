package storage_test

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/stretchr/testify/require"
)

// failingSecretsStore wraps a real secrets.Store, optionally forcing Create
// and/or Resolve to fail, to test how the sanitizer handles secret store
// errors.
type failingSecretsStore struct {
	secrets.Store
	createErr  error
	resolveErr error
}

func (f failingSecretsStore) Create(ctx context.Context, keyName, keyValue, value string) error {
	if f.createErr != nil {
		return f.createErr
	}
	return f.Store.Create(ctx, keyName, keyValue, value)
}

func (f failingSecretsStore) Resolve(ctx context.Context, keyName, keyValue string) (string, error) {
	if f.resolveErr != nil {
		return "", f.resolveErr
	}
	return f.Store.Resolve(ctx, keyName, keyValue)
}

func TestSanitizer_Parameters(t *testing.T) {
	c := portercontext.New()
	bun, err := cnab.LoadBundle(c, filepath.Join("../porter/testdata/bundle.json"))
	require.NoError(t, err)

	ctx := context.Background()
	r := porter.NewTestPorter(t)
	defer r.Close()

	recordID := "01FZVC5AVP8Z7A78CSCP1EJ604"
	sensitiveParamName := "my-second-param"
	sensitiveParamKey := recordID + "-" + sensitiveParamName
	expected := []secrets.SourceMap{
		{Name: "my-first-param", Source: secrets.Source{Strategy: host.SourceValue, Hint: "1"}, ResolvedValue: "1"},
		{Name: sensitiveParamName, Source: secrets.Source{Strategy: secrets.SourceSecret, Hint: sensitiveParamKey}, ResolvedValue: "2"},
	}
	sort.SliceStable(expected, func(i, j int) bool {
		return expected[i].Name < expected[j].Name
	})

	// parameters that are hard coded values should be sanitized, while those mapped from secrets or env vars should be left alone by the sanitizer
	rawParams := map[string]interface{}{
		"my-first-param":   1,
		sensitiveParamName: "2",
	}
	result, err := r.TestSanitizer.CleanRawParameters(ctx, rawParams, bun, recordID)
	require.NoError(t, err)
	require.Equal(t, len(expected), len(result))
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	require.Truef(t, reflect.DeepEqual(result, expected), "expected: %v, got: %v", expected, result)

	pset := storage.NewParameterSet("", "dev", result...)
	resolved, err := r.TestSanitizer.RestoreParameterSet(ctx, pset, bun)
	require.NoError(t, err)

	require.Equal(t, len(rawParams), len(resolved))
	for name, value := range resolved {
		require.Equal(t, rawParams[name], value)
	}
}

func TestSanitizer_CleanParameters(t *testing.T) {
	testcases := []struct {
		name       string
		paramName  string
		sourceKey  string
		wantSource secrets.Source
	}{
		{ // Should be switched to a secret
			name:       "hardcoded sensitive value",
			paramName:  "my-second-param",
			sourceKey:  host.SourceValue,
			wantSource: secrets.Source{Strategy: secrets.SourceSecret, Hint: "INSTALLATION_ID-my-second-param"},
		},
		{ // Should be left alone
			name:       "hardcoded insensitive value",
			paramName:  "my-first-param",
			sourceKey:  host.SourceValue,
			wantSource: secrets.Source{Strategy: host.SourceValue, Hint: "myvalue"},
		},
		{ // Should be left alone
			name:       "secret",
			paramName:  "my-first-param",
			sourceKey:  secrets.SourceSecret,
			wantSource: secrets.Source{Strategy: secrets.SourceSecret, Hint: "myvalue"},
		},
		{ // Should be left alone
			name:       "env var",
			paramName:  "my-second-param",
			sourceKey:  host.SourceEnv,
			wantSource: secrets.Source{Strategy: host.SourceEnv, Hint: "myvalue"},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			c := portercontext.New()
			bun, err := cnab.LoadBundle(c, filepath.Join("../porter/testdata/bundle.json"))
			require.NoError(t, err)

			ctx := context.Background()
			r := porter.NewTestPorter(t)
			defer r.Close()

			inst := storage.NewInstallation("", "mybuns")
			inst.ID = "INSTALLATION_ID" // Standardize for easy comparisons later
			inst.Parameters.Parameters = []secrets.SourceMap{
				{Name: tc.paramName, Source: secrets.Source{Strategy: tc.sourceKey, Hint: "myvalue"}},
			}
			gotParams, err := r.Sanitizer.CleanParameters(ctx, inst.Parameters.Parameters, bun, inst.ID)
			require.NoError(t, err, "CleanParameters failed")

			wantParms := []secrets.SourceMap{{Name: tc.paramName, Source: tc.wantSource}}
			require.Equal(t, wantParms, gotParams, "unexpected value returned from CleanParameters")
		})
	}
}

func TestSanitizer_Output(t *testing.T) {
	c := portercontext.New()
	bun, err := cnab.LoadBundle(c, filepath.Join("../porter/testdata/bundle.json"))
	require.NoError(t, err)

	ctx := context.Background()
	r := porter.NewTestPorter(t)
	defer r.Close()

	recordID := "01FZVC5AVP8Z7A78CSCP1EJ604"
	sensitiveOutputName := "my-first-output"
	sensitiveOutput := storage.Output{
		Name:  sensitiveOutputName,
		Key:   "",
		Value: []byte("this is secret output"),
		RunID: recordID,
	}

	expectedSensitiveOutput := storage.Output{
		Name:  sensitiveOutputName,
		Key:   recordID + "-" + sensitiveOutputName,
		Value: nil,
		RunID: recordID,
	}

	plainOutput := storage.Output{
		Name:  "my-second-output",
		Key:   "",
		Value: []byte("true"),
		RunID: recordID,
	}

	plainResult, err := r.TestSanitizer.CleanOutput(ctx, plainOutput, bun)
	require.NoError(t, err)
	require.Equal(t, plainOutput, plainResult)

	sensitiveResult, err := r.TestSanitizer.CleanOutput(ctx, sensitiveOutput, bun)
	require.NoError(t, err)
	require.Equal(t, expectedSensitiveOutput, sensitiveResult)

	expectedOutputs := storage.NewOutputs([]storage.Output{
		plainOutput,
		{Name: sensitiveOutputName, Key: expectedSensitiveOutput.Key, Value: sensitiveOutput.Value, RunID: recordID},
	})
	resolved, err := r.TestSanitizer.RestoreOutputs(ctx, storage.NewOutputs([]storage.Output{sensitiveResult, plainOutput}))
	require.NoError(t, err)
	sort.Sort(resolved)
	sort.Sort(expectedOutputs)
	require.Truef(t, reflect.DeepEqual(expectedOutputs, resolved), "expected outputs: %v, got outputs: %v", expectedOutputs, resolved)

}

func TestSanitizer_CleanOutput_SecretCreateFails(t *testing.T) {
	c := portercontext.New()
	bun, err := cnab.LoadBundle(c, filepath.Join("../porter/testdata/bundle.json"))
	require.NoError(t, err)

	ctx := context.Background()
	r := porter.NewTestPorter(t)
	defer r.Close()

	createErr := errors.New("secret store unreachable")
	failingSecrets := failingSecretsStore{Store: r.TestSecrets, createErr: createErr}
	sanitizer := storage.NewSanitizer(r.TestParameters, failingSecrets)

	recordID := "01FZVC5AVP8Z7A78CSCP1EJ604"
	sensitiveOutputName := "my-first-output"
	sensitiveOutput := storage.Output{
		Name:  sensitiveOutputName,
		Value: []byte("this is secret output"),
		RunID: recordID,
	}

	result, err := sanitizer.CleanOutput(ctx, sensitiveOutput, bun)
	require.ErrorIs(t, err, createErr)
	require.Empty(t, result.Key, "no Key should be persisted when the secret was never created")
	require.Empty(t, result.Value, "the raw sensitive value should never be persisted to the primary store")
}

func TestSanitizer_RestoreOutputs_TolerantOfResolveFailure(t *testing.T) {
	ctx := context.Background()
	r := porter.NewTestPorter(t)
	defer r.Close()

	recordID := "01FZVC5AVP8Z7A78CSCP1EJ604"
	plainOutput := storage.Output{
		Name:  "my-second-output",
		Value: []byte("true"),
		RunID: recordID,
	}

	// Simulate a dangling reference: a Key that points at a secret that
	// doesn't exist (e.g. never created, or later deleted).
	danglingOutput := storage.Output{
		Name:  "my-first-output",
		Key:   recordID + "-my-first-output",
		RunID: recordID,
	}

	outputs, err := r.TestSanitizer.RestoreOutputs(ctx, storage.NewOutputs([]storage.Output{plainOutput, danglingOutput}))
	require.NoError(t, err, "RestoreOutputs should not fail the whole batch when one output can't be resolved")

	restoredPlain, ok := outputs.GetByName(plainOutput.Name)
	require.True(t, ok)
	require.Equal(t, plainOutput.Value, restoredPlain.Value, "unaffected outputs should still resolve normally")

	restoredDangling, ok := outputs.GetByName(danglingOutput.Name)
	require.True(t, ok)
	require.Empty(t, restoredDangling.Value)
	require.NotEmpty(t, restoredDangling.PersistError, "the unresolvable output should be marked with a PersistError instead of failing the batch")
}
