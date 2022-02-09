package claims

import (
	"context"
	"encoding/base64"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	cnabclaims "github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ Provider = ClaimStore{}

var b64encode = func(src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(dst, src)
	return dst, nil
}

var b64decode = func(src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))
	n, err := base64.StdEncoding.Decode(dst, src)
	return dst[:n], err
}

var exampleBundle = bundle.Bundle{
	SchemaVersion:    "schemaVersion",
	Name:             "mybun",
	Version:          "v0.1.0",
	Description:      "this is my bundle",
	InvocationImages: []bundle.InvocationImage{},
	Actions: map[string]bundle.Action{
		"test": {Modifies: true},
		"logs": {Modifies: false},
	},
}

// generateClaimData creates test claims, results and outputs
// it returns a claim Provider, and a test cleanup function.
//
// claims/
//   foo/
//     CLAIM_ID_1 (install)
//     CLAIM_ID_2 (upgrade)
//     CLAIM_ID_3 (invoke - test)
//     CLAIM_ID_4 (uninstall)
//   bar/
//     CLAIM_ID_10 (install)
//   baz/
//     CLAIM_ID_20 (install)
//     CLAIM_ID_21 (install)
// results/
//   CLAIM_ID_1/
//     RESULT_ID_1 (success)
//   CLAIM_ID_2/
//     RESULT_ID 2 (success)
//   CLAIM_ID_3/
//     RESULT_ID_3 (failed)
//   CLAIM_ID_4/
//     RESULT_ID_4 (success)
//   CLAIM_ID_10/
//     RESULT_ID_10 (running)
//     RESULT_ID_11 (success)
//   CLAIM_ID_20/
//     RESULT_ID_20 (failed)
//   CLAIM_ID_21/
//     NO RESULT YET
// outputs/
//   RESULT_ID_1/
//     RESULT_ID_1_OUTPUT_1
//   RESULT_ID_2/
//     RESULT_ID_2_OUTPUT_1
//     RESULT_ID_2_OUTPUT_2
func generateClaimData(t *testing.T) *TestClaimProvider {
	cp := NewTestClaimProvider(t)

	bun := bundle.Bundle{
		Definitions: map[string]*definition.Schema{
			"output1": {
				Type: "string",
			},
			"output2": {
				Type: "string",
			},
		},
		Outputs: map[string]bundle.Output{
			"output1": {
				Definition: "output1",
			},
			"output2": {
				Definition: "output2",
				ApplyTo:    []string{"upgrade"},
			},
		},
	}

	setBun := func(r *Run) { r.Bundle = bun }

	// Create the foo installation data
	foo := cp.CreateInstallation(NewInstallation("dev", "foo"))
	run := cp.CreateRun(foo.NewRun(cnab.ActionInstall), setBun)
	result := cp.CreateResult(run.NewResult(cnab.StatusSucceeded))
	cp.CreateOutput(result.NewOutput(cnabclaims.OutputInvocationImageLogs, []byte("install logs")))
	cp.CreateOutput(result.NewOutput("output1", []byte("install output1")))

	run = cp.CreateRun(foo.NewRun(cnab.ActionUpgrade), setBun)
	result = cp.CreateResult(run.NewResult(cnab.StatusSucceeded))
	cp.CreateOutput(result.NewOutput(cnabclaims.OutputInvocationImageLogs, []byte("upgrade logs")))
	cp.CreateOutput(result.NewOutput("output1", []byte("upgrade output1")))
	cp.CreateOutput(result.NewOutput("output2", []byte("upgrade output2")))
	// Test bug in how we read output names by having the name include characters from the result id
	cp.CreateOutput(result.NewOutput(result.ID+"-output3", []byte("upgrade output3")))

	run = cp.CreateRun(foo.NewRun("test"), setBun)
	result = cp.CreateResult(run.NewResult(cnab.StatusFailed))

	run = cp.CreateRun(foo.NewRun(cnab.ActionUninstall), setBun)
	result = cp.CreateResult(run.NewResult(cnab.StatusSucceeded))

	// Record the status of the foo installation
	foo.ApplyResult(run, result)
	require.NoError(t, cp.UpdateInstallation(foo))

	// Create the bar installation data
	bar := cp.CreateInstallation(NewInstallation("dev", "bar"))
	run = cp.CreateRun(bar.NewRun(cnab.ActionInstall), setBun)
	cp.CreateResult(run.NewResult(cnab.StatusRunning))
	result = cp.CreateResult(run.NewResult(cnab.StatusSucceeded))

	// Record the status of the bar installation
	bar.ApplyResult(run, result)
	require.NoError(t, cp.UpdateInstallation(bar))

	// Create the baz installation data
	baz := cp.CreateInstallation(NewInstallation("dev", "baz"))
	run = cp.CreateRun(baz.NewRun(cnab.ActionInstall), setBun)
	cp.CreateResult(run.NewResult(cnab.StatusFailed))
	run = cp.CreateRun(baz.NewRun(cnab.ActionInstall), setBun)
	result = cp.CreateResult(run.NewResult(cnab.StatusRunning))

	// Record the status of the baz installation
	baz.ApplyResult(run, result)
	require.NoError(t, cp.UpdateInstallation(baz))

	return cp
}

func TestClaimStore_Installations(t *testing.T) {
	cp := generateClaimData(t)
	defer cp.Teardown()

	t.Run("ListInstallations", func(t *testing.T) {
		installations, err := cp.ListInstallations(context.Background(), "dev", "", nil)
		require.NoError(t, err, "ListInstallations failed")

		require.Len(t, installations, 3, "Expected 3 installations")

		bar := installations[0]
		assert.Equal(t, "bar", bar.Name)
		assert.Equal(t, cnab.StatusSucceeded, bar.Status.ResultStatus)

		baz := installations[1]
		assert.Equal(t, "baz", baz.Name)
		assert.Equal(t, cnab.StatusRunning, baz.Status.ResultStatus)

		foo := installations[2]
		assert.Equal(t, "foo", foo.Name)
		assert.Equal(t, cnab.StatusSucceeded, foo.Status.ResultStatus)
	})

	t.Run("GetInstallation", func(t *testing.T) {
		foo, err := cp.GetInstallation("dev", "foo")
		require.NoError(t, err, "GetInstallation failed")

		assert.Equal(t, "foo", foo.Name)
		assert.Equal(t, cnab.StatusSucceeded, foo.Status.ResultStatus)
	})

	t.Run("GetInstallation - not found", func(t *testing.T) {
		_, err := cp.GetInstallation("", "missing")
		require.ErrorIs(t, err, storage.ErrNotFound{})
	})

}

func TestClaimStore_DeleteInstallation(t *testing.T) {
	cp := generateClaimData(t)
	defer cp.Teardown()

	installations, err := cp.ListInstallations(context.Background(), "dev", "", nil)
	require.NoError(t, err, "ListInstallations failed")
	assert.Len(t, installations, 3, "expected 3 installations")

	err = cp.RemoveInstallation("dev", "foo")
	require.NoError(t, err, "RemoveInstallation failed")

	installations, err = cp.ListInstallations(context.Background(), "dev", "", nil)
	require.NoError(t, err, "ListInstallations failed")
	assert.Len(t, installations, 2, "expected foo to be deleted")

	_, err = cp.GetLastRun("dev", "foo")
	require.ErrorIs(t, err, storage.ErrNotFound{})
}

func TestClaimStore_Run(t *testing.T) {
	cp := generateClaimData(t)

	t.Run("ListRuns", func(t *testing.T) {
		runs, resultsMap, err := cp.ListRuns("dev", "foo")
		require.NoError(t, err, "Failed to read claims: %s", err)

		require.Len(t, runs, 4, "Expected 4 runs")
		require.Len(t, resultsMap, 4, "Results expected to have 4 runs")
		assert.Equal(t, cnab.ActionInstall, runs[0].Action)
		assert.Equal(t, cnab.ActionUpgrade, runs[1].Action)
		assert.Equal(t, "test", runs[2].Action)
		assert.Equal(t, cnab.ActionUninstall, runs[3].Action)
	})

	t.Run("ListRuns - bundle not yet run", func(t *testing.T) {
		// It's now possible for someone to create an installation and not immediately have any runs.
		runs, resultsMap, err := cp.ListRuns("dev", "missing")
		require.NoError(t, err)
		assert.Empty(t, runs)
		assert.Empty(t, resultsMap)
	})

	t.Run("GetRun", func(t *testing.T) {
		runs, _, err := cp.ListRuns("dev", "foo")
		require.NoError(t, err, "ListRuns failed")

		assert.NotEmpty(t, runs, "no claims were found")
		runID := runs[0].ID

		c, err := cp.GetRun(runID)
		require.NoError(t, err, "GetRun failed")

		assert.Equal(t, "foo", c.Installation)
		assert.Equal(t, cnab.ActionInstall, c.Action)
	})

	t.Run("GetRun - invalid claim", func(t *testing.T) {
		_, err := cp.GetRun("missing")
		require.ErrorIs(t, err, storage.ErrNotFound{})
	})

	t.Run("GetLastRun", func(t *testing.T) {
		c, err := cp.GetLastRun("dev", "bar")
		require.NoError(t, err, "GetLastRun failed")

		assert.Equal(t, "bar", c.Installation)
		assert.Equal(t, cnab.ActionInstall, c.Action)
	})

	t.Run("GetLastRun - invalid installation", func(t *testing.T) {
		_, err := cp.GetLastRun("dev", "missing")
		require.ErrorIs(t, err, storage.ErrNotFound{})
	})
}

func TestClaimStore_Results(t *testing.T) {
	cp := generateClaimData(t)
	defer cp.Teardown()

	barRuns, resultsMap, err := cp.ListRuns("dev", "bar")
	require.NoError(t, err, "ListRuns failed")
	require.Len(t, barRuns, 1, "expected 1 claim")
	require.Len(t, resultsMap, 1, "expected 1 claim")
	runID := barRuns[0].ID // this claim has multiple results

	t.Run("ListResults", func(t *testing.T) {
		results, err := cp.ListResults(runID)
		require.NoError(t, err, "ListResults failed")
		assert.Len(t, results, 2, "expected 2 results")
		assert.Len(t, resultsMap[runID], 2, "expected 2 results for runID in results map")
	})

	t.Run("GetResult", func(t *testing.T) {
		results, err := cp.ListResults(runID)
		require.NoError(t, err, "ListResults failed")

		resultID := results[0].ID

		r, err := cp.GetResult(resultID)
		require.NoError(t, err, "GetResult failed")

		assert.Equal(t, cnab.StatusRunning, r.Status)
	})

	t.Run("ReadResult - invalid result", func(t *testing.T) {
		_, err := cp.GetResult("missing")
		require.ErrorIs(t, err, storage.ErrNotFound{})
	})
}

func TestClaimStore_Outputs(t *testing.T) {
	cp := generateClaimData(t)
	defer cp.Teardown()

	fooRuns, _, err := cp.ListRuns("dev", "foo")
	require.NoError(t, err, "ListRuns failed")
	require.NotEmpty(t, fooRuns, "expected foo to have a run")
	foo := fooRuns[1]
	fooResults, err := cp.ListResults(foo.ID) // Use foo's upgrade claim that has two outputs
	require.NoError(t, err, "ListResults failed")
	require.NotEmpty(t, fooResults, "expected foo to have a result")
	fooResult := fooResults[0]
	resultID := fooResult.ID // this result has an output

	barRuns, _, err := cp.ListRuns("dev", "bar")
	require.NoError(t, err, "ReadAllClaims failed")
	require.Len(t, barRuns, 1, "expected bar to have a run")
	barRun := barRuns[0]
	barResults, err := cp.ListResults(barRun.ID)
	require.NoError(t, err, "ReadAllResults failed")
	require.NotEmpty(t, barResults, "expected bar to have a result")
	barResult := barResults[0]
	resultIDWithoutOutputs := barResult.ID

	t.Run("ListOutputs", func(t *testing.T) {
		outputs, err := cp.ListOutputs(resultID)
		require.NoError(t, err, "ListResults failed")
		assert.Len(t, outputs, 4, "expected 2 outputs")

		assert.Equal(t, outputs[0].Name, resultID+"-output3")
		assert.Equal(t, outputs[1].Name, cnabclaims.OutputInvocationImageLogs)
		assert.Equal(t, outputs[2].Name, "output1")
		assert.Equal(t, outputs[3].Name, "output2")
	})

	t.Run("ListOutputs - no outputs", func(t *testing.T) {
		outputs, err := cp.ListResults(resultIDWithoutOutputs)
		require.NoError(t, err, "listing outputs for a result that doesn't have any should not result in an error")
		assert.Empty(t, outputs)
	})

	t.Run("GetLastOutputs", func(t *testing.T) {
		outputs, err := cp.GetLastOutputs("dev", "foo")

		require.NoError(t, err, "GetLastOutputs failed")
		assert.Equal(t, 4, outputs.Len(), "wrong number of outputs identified")

		gotOutput1, hasOutput1 := outputs.GetByName("output1")
		assert.True(t, hasOutput1, "should have found output1")
		assert.Equal(t, "upgrade output1", string(gotOutput1.Value), "did not find the most recent value for output1")

		gotOutput2, hasOutput2 := outputs.GetByName("output2")
		assert.True(t, hasOutput2, "should have found output2")
		assert.Equal(t, "upgrade output2", string(gotOutput2.Value), "did not find the most recent value for output2")
	})

	t.Run("ReadLastOutputs - invalid installation", func(t *testing.T) {
		outputs, err := cp.GetLastOutputs("dev", "missing")
		require.NoError(t, err)
		assert.Equal(t, outputs.Len(), 0)
	})

	t.Run("GetLastOutput", func(t *testing.T) {
		o, err := cp.GetLastOutput("dev", "foo", "output1")

		require.NoError(t, err, "GetLastOutputs failed")
		assert.Equal(t, "upgrade output1", string(o.Value), "did not find the most recent value for output1")

	})

	t.Run("GetLastOutput - invalid installation", func(t *testing.T) {
		o, err := cp.GetLastOutput("dev", "missing", "output1")
		require.ErrorIs(t, err, storage.ErrNotFound{})
		assert.Empty(t, o)
	})

	t.Run("GetLastLogs", func(t *testing.T) {
		logs, hasLogs, err := cp.GetLastLogs("dev", "foo")

		require.NoError(t, err, "GetLastLogs failed")
		assert.True(t, hasLogs, "expected logs to be found")
		assert.Equal(t, "upgrade logs", logs, "did not find the most recent logs for foo")
	})
}
