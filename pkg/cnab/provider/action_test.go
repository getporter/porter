package cnabprovider

import (
	"context"
	"encoding/json"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/test"
	"os"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddRelocation(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/relocation-mapping.json")
	require.NoError(t, err)

	d := NewTestRuntime(t)
	defer d.Close()

	var args ActionArguments
	require.NoError(t, json.Unmarshal(data, &args.BundleReference.RelocationMap))

	opConf := d.AddRelocation(args)

	invoImage := bundle.InvocationImage{}
	invoImage.Image = "gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687"

	op := &driver.Operation{
		Files: make(map[string]string),
		Image: invoImage,
	}
	err = opConf(op)
	assert.NoError(t, err)

	mapping, ok := op.Files["/cnab/app/relocation-mapping.json"]
	assert.True(t, ok)
	assert.Equal(t, string(data), mapping)
	assert.Equal(t, "my.registry/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687", op.Image.Image)

}

func TestAddFiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d := NewTestRuntime(t)
	defer d.Close()

	// Make a test claim
	instName := "mybuns"
	run1 := storage.NewRun("", instName)
	run1.NewResult(cnab.StatusPending)
	i := d.TestInstallations.CreateInstallation(storage.NewInstallation("", instName), d.TestInstallations.SetMutableInstallationValues)
	d.TestInstallations.CreateRun(run1, d.TestInstallations.SetMutableRunValues)

	// Prep the files in the bundle
	args := ActionArguments{
		Installation: i,
	}
	op := &driver.Operation{}
	err := d.AddFiles(ctx, args)(op)
	require.NoError(t, err, "AddFiles failed")

	// Check that we injected a CNAB claim and not our Run representation, they aren't exactly 1:1 the same format
	require.Contains(t, op.Files, config.ClaimFilepath, "The claim should have been injected into the bundle")
	test.CompareGoldenFile(t, "testdata/want-claim.json", op.Files[config.ClaimFilepath])
}

func TestSaveOperationResult_ModifiesFalse_SkipsPorterState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d := NewTestRuntime(t)
	defer d.Close()

	instName := "mybuns"
	bun := bundle.Bundle{
		Actions: map[string]bundle.Action{
			"dry-run": {Modifies: false},
		},
		Outputs: map[string]bundle.Output{
			"porter-state": {Definition: "porter-state", Path: "/cnab/app/outputs/porter-state.tgz"},
			"user-output":  {Definition: "user-output", Path: "/cnab/app/outputs/user-output"},
		},
		Definitions: map[string]*definition.Schema{
			"porter-state": {Comment: cnab.PorterInternal},
			"user-output":  {Type: "string"},
		},
	}
	i := d.TestInstallations.CreateInstallation(storage.NewInstallation("", instName), d.TestInstallations.SetMutableInstallationValues)
	run := storage.NewRun("", instName)
	run.Bundle = bun
	run.Action = "dry-run"
	run = d.TestInstallations.CreateRun(run, d.TestInstallations.SetMutableRunValues)
	result := run.NewResult(cnab.StatusSucceeded)

	opResult := driver.OperationResult{
		Outputs: map[string]string{
			"porter-state": "state-data",
			"user-output":  "hello",
		},
	}

	err := d.SaveOperationResult(ctx, opResult, i, run, result)
	require.NoError(t, err)

	outputs, err := d.TestInstallations.GetOutputs(ctx, run.ID)
	require.NoError(t, err)

	_, hasPorterState := outputs.GetByName("porter-state")
	assert.False(t, hasPorterState, "porter-state should not be saved for modifies:false actions")

	_, hasUserOutput := outputs.GetByName("user-output")
	assert.True(t, hasUserOutput, "user-defined outputs should still be saved")
}

func TestSaveOperationResult_ModifiesTrue_SavesPorterState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d := NewTestRuntime(t)
	defer d.Close()

	instName := "mybuns"
	bun := bundle.Bundle{
		Actions: map[string]bundle.Action{
			"rotate-creds": {Modifies: true},
		},
		Outputs: map[string]bundle.Output{
			"porter-state": {Definition: "porter-state", Path: "/cnab/app/outputs/porter-state.tgz"},
		},
		Definitions: map[string]*definition.Schema{
			"porter-state": {Comment: cnab.PorterInternal},
		},
	}
	i := d.TestInstallations.CreateInstallation(storage.NewInstallation("", instName), d.TestInstallations.SetMutableInstallationValues)
	run := storage.NewRun("", instName)
	run.Bundle = bun
	run.Action = "rotate-creds"
	run = d.TestInstallations.CreateRun(run, d.TestInstallations.SetMutableRunValues)
	result := run.NewResult(cnab.StatusSucceeded)

	opResult := driver.OperationResult{
		Outputs: map[string]string{
			"porter-state": "state-data",
		},
	}

	err := d.SaveOperationResult(ctx, opResult, i, run, result)
	require.NoError(t, err)

	outputs, err := d.TestInstallations.GetOutputs(ctx, run.ID)
	require.NoError(t, err)

	_, hasPorterState := outputs.GetByName("porter-state")
	assert.True(t, hasPorterState, "porter-state should be saved for modifies:true actions")
}
