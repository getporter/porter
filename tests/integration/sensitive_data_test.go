//go:build integration

package integration

import (
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensitiveData(t *testing.T) {
	t.Parallel()

	p := porter.NewTestPorter(t)
	defer p.Close()
	ctx := p.SetupIntegrationTest()

	bundleName := p.AddTestBundleDir("testdata/bundles/bundle-with-sensitive-data", true)

	sensitiveParamName := "password"
	sensitiveParamValue := "secretpassword"
	installOpts := porter.NewInstallOptions()
	installOpts.Params = []string{sensitiveParamName + "=" + sensitiveParamValue, "name=porter-test"}

	err := installOpts.Validate(ctx, []string{}, p.Porter)
	require.NoError(t, err)

	err = p.InstallBundle(ctx, installOpts)
	require.NoError(t, err)

	i, err := p.Installations.GetInstallation(ctx, installOpts.Namespace, installOpts.Name)
	require.NoError(t, err)

	run, err := p.Installations.GetRun(ctx, i.Status.RunID)
	require.NoError(t, err)

	sensitiveParam, ok := i.Parameters.Get(sensitiveParamName)
	require.True(t, ok)
	assert.NotContains(t, sensitiveParam.Source.Hint, sensitiveParamValue)
	assert.NotContains(t, sensitiveParam.Source.Hint, sensitiveParamValue)
	sensitiveOverride, ok := run.ParameterOverrides.Get(sensitiveParamName)
	require.True(t, ok)
	assert.NotContains(t, sensitiveOverride.Source.Hint, sensitiveParamValue)

	outputs, err := p.Installations.GetLastOutputs(ctx, "", bundleName)
	require.NoError(t, err, "GetLastOutput failed")
	mylogs, ok := outputs.GetByName("mylogs")
	require.True(t, ok, "expected mylogs output to be persisted")
	assert.Contains(t, string(mylogs.Value), "porter-test")
	result, ok := outputs.GetByName("result")
	require.True(t, ok, "expected result output to be persisted")
	assert.NotContains(t, string(result.Value), sensitiveParamValue)
}
