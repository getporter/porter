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
