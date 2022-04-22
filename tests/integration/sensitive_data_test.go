//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensitiveData(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Teardown()
	p.SetupIntegrationTest()
	p.Debug = false

	bundleName := p.AddTestBundleDir("testdata//bundles/bundle-with-sensitive-data", true)

	sensitiveParamValue := "secretpassword"
	installOpts := porter.NewInstallOptions()
	installOpts.Params = []string{"password=" + sensitiveParamValue, "name=porter-test"}

	err := installOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(context.Background(), installOpts)
	require.NoError(t, err)

	i, err := p.Claims.GetInstallation(installOpts.Namespace, installOpts.Name)
	require.NoError(t, err)

	run, err := p.Claims.GetRun(i.Status.RunID)
	require.NoError(t, err)

	bun := cnab.ExtendedBundle{run.Bundle}

	for _, param := range i.Parameters.Parameters {
		if bun.IsSensitiveParameter(param.Name) {
			assert.NotContains(t, param.Source.Value, sensitiveParamValue)
		}
	}

	for _, param := range run.ParameterOverrides.Parameters {
		if bun.IsSensitiveParameter(param.Name) {
			assert.NotContains(t, param.Source.Value, sensitiveParamValue)
		}
	}
	for _, param := range run.Parameters.Parameters {
		if bun.IsSensitiveParameter(param.Name) {
			assert.NotContains(t, param.Source.Value, sensitiveParamValue)
		}
	}

	outputs, err := p.Claims.GetLastOutputs("", bundleName)
	require.NoError(t, err, "GetLastOutput failed")
	mylogs, ok := outputs.GetByName("mylogs")
	require.True(t, ok, "expected mylogs output to be persisted")
	assert.Contains(t, string(mylogs.Value), "porter-test")
	result, ok := outputs.GetByName("result")
	require.True(t, ok, "expected result output to be persisted")
	assert.NotContains(t, string(result.Value), sensitiveParamValue)
}
