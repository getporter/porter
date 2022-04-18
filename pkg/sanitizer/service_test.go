package sanitizer_test

import (
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/parameters"
	"get.porter.sh/porter/pkg/porter"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"github.com/stretchr/testify/require"
)

func TestSanitizer_Parameters(t *testing.T) {
	cxt := portercontext.New()
	bun, err := cnab.LoadBundle(cxt, filepath.Join("../porter/testdata/bundle.json"))
	require.NoError(t, err)

	r := porter.NewTestPorter(t)
	defer r.Teardown()

	recordID := "01FZVC5AVP8Z7A78CSCP1EJ604"
	sensitiveParamName := "my-second-param"
	sensitiveParamKey := recordID + sensitiveParamName
	expected := []secrets.Strategy{
		{Name: "my-first-param", Source: secrets.Source{Key: parameters.DEFAULT_STRATEGY, Value: "1"}, Value: "1"},
		{Name: sensitiveParamName, Source: secrets.Source{Key: secrets.SourceSecret, Value: sensitiveParamKey}, Value: "2"},
	}
	sort.SliceStable(expected, func(i, j int) bool {
		return expected[i].Name < expected[j].Name
	})

	rawParams := map[string]interface{}{
		"my-first-param":   1,
		sensitiveParamName: "2",
	}
	result, err := r.TestSanitizer.CleanRawParameters(rawParams, bun, recordID)
	require.NoError(t, err)
	require.Equal(t, len(expected), len(result))
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	require.Truef(t, reflect.DeepEqual(result, expected), "expected: %v, got: %v", expected, result)

	pset := parameters.NewParameterSet("", "dev", result...)
	resolved, err := r.TestSanitizer.RestoreParameterSet(pset, bun)
	require.NoError(t, err)

	require.Equal(t, len(rawParams), len(resolved))
	for name, value := range resolved {
		require.Equal(t, rawParams[name], value)
	}
}

func TestSanitizer_Output(t *testing.T) {
	cxt := portercontext.New()
	bun, err := cnab.LoadBundle(cxt, filepath.Join("../porter/testdata/bundle.json"))
	require.NoError(t, err)

	r := porter.NewTestPorter(t)
	defer r.Teardown()

	recordID := "01FZVC5AVP8Z7A78CSCP1EJ604"
	sensitiveOutputName := "my-first-output"
	sensitiveOutput := claims.Output{
		Name:  sensitiveOutputName,
		Key:   "",
		Value: []byte("this is secret output"),
		RunID: recordID,
	}

	expectedSensitiveOutput := claims.Output{
		Name:  sensitiveOutputName,
		Key:   recordID + sensitiveOutputName,
		Value: []byte{},
		RunID: recordID,
	}

	plainOutput := claims.Output{
		Name:  "my-second-output",
		Key:   "",
		Value: []byte("true"),
		RunID: recordID,
	}

	plainResult, err := r.TestSanitizer.CleanOutput(plainOutput, bun)
	require.NoError(t, err)
	require.Equal(t, plainOutput, plainResult)

	sensitiveResult, err := r.TestSanitizer.CleanOutput(sensitiveOutput, bun)
	require.NoError(t, err)
	require.Equal(t, expectedSensitiveOutput, sensitiveResult)

	expectedOutputs := claims.NewOutputs([]claims.Output{
		plainOutput,
		{Name: sensitiveOutputName, Key: expectedSensitiveOutput.Key, Value: sensitiveOutput.Value, RunID: recordID},
	})
	resolved, err := r.TestSanitizer.RestoreOutputs(claims.NewOutputs([]claims.Output{sensitiveResult, plainOutput}))
	require.NoError(t, err)
	sort.Sort(resolved)
	sort.Sort(expectedOutputs)
	require.Truef(t, reflect.DeepEqual(expectedOutputs, resolved), "expected outputs: %v, got outputs: %v", expectedOutputs, resolved)

}
