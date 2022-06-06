package migrations

import (
	"io/ioutil"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertInstallation(t *testing.T) {
	inst := convertInstallation("mybuns")
	assert.NotEmpty(t, inst.ID, "an id was not assigned")
	assert.Empty(t, inst.Namespace, "by default installations are migrated into the global namespace")
	assert.Equal(t, "mybuns", inst.Name, "incorrect name")
	assert.Equal(t, storage.InstallationSchemaVersion, inst.SchemaVersion, "incorrect schema version")
}

func TestConvertClaimToRun(t *testing.T) {
	inst := storage.NewInstallation("", "mybuns")

	claimData, err := ioutil.ReadFile("testdata/v0_home/claims/hello1/01G1VJGY43HT3KZN82DS6DDPWK.json")
	require.NoError(t, err, "could not read testdata")

	run, err := convertClaimToRun(inst, claimData)
	require.NoError(t, err, "error converting claim")

	assert.Equal(t, "01G1VJGY43HT3KZN82DS6DDPWK", run.ID, "incorrect run id")
	assert.Equal(t, storage.InstallationSchemaVersion, run.SchemaVersion, "incorrect schema version, should be the current schema supported by porter")
	assert.Equal(t, "hello1", run.Installation, "incorrect installation name")
	assert.Equal(t, "01G1VJGY43HT3KZN82DSJY4NNB", run.Revision, "incorrect revision")
	assert.Equal(t, "2022-04-29T16:09:42.65907-05:00", run.Created.Format(time.RFC3339Nano), "incorrect created timestamp")
	assert.Equal(t, "install", run.Action, "incorrect action")
	assert.NotEmpty(t, run.Bundle, "bundle was not populated")
	assert.Len(t, run.Parameters.Parameters, 1, "incorrect parameters")

	param := run.Parameters.Parameters[0]
	assert.Equal(t, param.Name, "porter-debug", "incorrect parameter name")
	assert.Equal(t, param.Source.Key, "value", "incorrect parameter source key")
	assert.Equal(t, param.Source.Value, "false", "incorrect parameter source value")
}

func TestConvertResult(t *testing.T) {
	run := storage.NewRun("myns", "mybuns")

	resultData, err := ioutil.ReadFile("testdata/v0_home/results/01G1VJGY43HT3KZN82DS6DDPWK/01G1VJH2HP97B5B0N5S37KYMVG.json")
	require.NoError(t, err, "could not read testdata result")

	result, err := convertResult(run, resultData)
	require.NoError(t, err, "failed to convert result")

	assert.Equal(t, storage.InstallationSchemaVersion, result.SchemaVersion, "incorrect schema version")
	assert.Equal(t, run.Namespace, result.Namespace, "incorrect namespace")
	assert.Equal(t, run.Installation, result.Installation, "incorrect installation name")
	assert.Equal(t, "yay!", result.Message, "incorrect message")
	assert.Equal(t, "boop!", result.Custom, "incorrect custom data")
	assert.Equal(t, "01G1VJH2HP97B5B0N5S37KYMVG", result.ID, "incorrect id")
	assert.Equal(t, run.ID, result.RunID, "incorrect run id")
	assert.Equal(t, "2022-04-29T16:09:47.190534-05:00", result.Created.Format(time.RFC3339Nano), "incorrect created timestamp")
	assert.Equal(t, "succeeded", result.Status, "incorrect status")
	assert.Len(t, result.OutputMetadata, 1, "expected one output populated")

	contentDigest, _ := result.OutputMetadata.GetContentDigest(cnab.OutputInvocationImageLogs)
	assert.Equal(t, "sha256:28ccd0529aa1edefb0e771a28c31c0193f656718af985fed197235ba98fc5696", contentDigest, "incorrect output content digest")

	generatedFlag, _ := result.OutputMetadata.GetGeneratedByBundle(cnab.OutputInvocationImageLogs)
	assert.False(t, generatedFlag, "incorrect generated output flag")
}

func TestConvertOutput(t *testing.T) {
	outputData, err := ioutil.ReadFile("testdata/v0_home/outputs/01G1VJH2HP97B5B0N5S37KYMVG/01G1VJH2HP97B5B0N5S37KYMVG-io.cnab.outputs.invocationImageLogs")
	require.NoError(t, err, "error reading testdata output file")

	run := storage.NewRun("myns", "mybuns")
	result := run.NewResult("succeeded")

	output, err := convertOutput(result, "01G1VJH2HP97B5B0N5S37KYMVG-io.cnab.outputs.invocationImageLogs", outputData)
	require.NoError(t, err, "error converting output")

	require.Equal(t, storage.InstallationSchemaVersion, output.SchemaVersion, "incorrect schema version")
	require.Equal(t, result.Namespace, output.Namespace, "incorrect namespace")
	require.Equal(t, result.Installation, output.Installation, "incorrect installation")
	require.Equal(t, "io.cnab.outputs.invocationImageLogs", output.Name, "incorrect name")
	require.Equal(t, result.ID, output.ResultID, "incorrect result id")
	require.Equal(t, result.RunID, output.RunID, "incorrect run id")
	require.Equal(t, outputData, output.Value, "incorrect output value")
}

func TestMigration_MigrateInstallations(t *testing.T) {
	c := createLegacyPorterHome(t)
	defer c.Close()
	home, err := c.GetHomeDir()
	require.NoError(t, err, "could not get the home directory")
	ctx, _, _ := c.SetupIntegrationTest()

	destStore := storage.NewTestStore(c)
	testSecrets := secrets.NewTestSecretsProvider()
	testParams := storage.NewTestParameterProviderFor(t, destStore, testSecrets)
	testSanitizer := storage.NewSanitizer(testParams, testSecrets)

	opts := storage.MigrateOptions{
		OldHome:              home,
		SourceAccount:        "src",
		DestinationNamespace: "myns",
	}
	m := NewMigration(c.Config, opts, destStore, testSanitizer)
	defer m.Close()

	err = m.Connect(ctx)
	require.NoError(t, err, "connect failed")

	err = m.migrateInstallations(ctx)
	require.NoError(t, err, "migrate installations failed")

	is := storage.NewInstallationStore(destStore)
	installations, err := is.ListInstallations(ctx, opts.DestinationNamespace, "", nil)
	require.NoError(t, err, "could not list installations in the destination database")
	assert.Len(t, installations, 1, "expected 1 installation to be migrated")

	// Validate that the installation as migrated correctly
	inst := installations[0]
	assert.Equal(t, storage.InstallationSchemaVersion, inst.SchemaVersion, "incorrect installation schema")
	assert.NotEmpty(t, inst.ID, "an installation id should have been assigned")
	assert.Equal(t, "hello1", inst.Name, "incorrect installation name")
	assert.Equal(t, opts.DestinationNamespace, inst.Namespace, "installation namespace should be set to the destination namespace")
	assert.Empty(t, inst.Bundle, "We didn't track the bundle reference in v0, so this can't be populated")
	assert.Empty(t, inst.Custom, "We didn't allow setting custom metadata on installations in v0, so this can't be populated")
	assert.Empty(t, inst.Labels, "We didn't allow setting labels on installations in v0, so this can't be populated")
	assert.Empty(t, inst.CredentialSets, "We didn't track credential sets used when running a bundle in v0, so this can't be populated")
	assert.Empty(t, inst.ParameterSets, "We didn't track parameter sets used when running a bundle in v0, so this can't be populated")
	assert.Empty(t, inst.Parameters.Parameters, "We didn't track manually specified parameters when running a bundle in v0, so this can't be populated")

	// Validate the installation status, which is calculated based on the runs and their results
	assert.Equal(t, "2022-04-28T16:09:42.65907-05:00", inst.Status.Created.Format(time.RFC3339Nano), "Created timestamp should be set to the timestamp of the first run")
	assert.Equal(t, "2022-04-29T16:13:20.48026-05:00", inst.Status.Modified.Format(time.RFC3339Nano), "Modified timestamp should be set to the timestamp of the last run")
	require.NotNil(t, inst.Status.Installed, "the install timestamp should be set")
	assert.Equal(t, "2022-04-29T16:09:47.190534-05:00", inst.Status.Installed.Format(time.RFC3339Nano), "the install timestamp should be set to the timestamp of the successful install result")
	assert.NotNil(t, inst.Status.Uninstalled, "the uninstall timestamp should be set")
	assert.Equal(t, "2022-04-29T16:13:21.802457-05:00", inst.Status.Uninstalled.Format(time.RFC3339Nano), "the uninstalled timestamp should be set to the timestamp of the successful uninstall result")
	assert.Equal(t, "01G1VJQJV0RN5AW5BSZHNTVYTV", inst.Status.RunID, "incorrect last run id set on the installation status")
	assert.Equal(t, "01G1VJQM4AVN7SCXC8WV3M0D7N", inst.Status.ResultID, "incorrect last result id set on the installation status")
	assert.Equal(t, "succeeded", inst.Status.ResultStatus, "the installation status should be successful")
	assert.Equal(t, "0.1.1", inst.Status.BundleVersion, "incorrect installation bundle version")
	assert.Empty(t, inst.Status.BundleReference, "We didn't track bundle reference in v0, so this can't be populated")
	assert.Empty(t, inst.Status.BundleDigest, "We didn't track bundle digest in v0, so this can't be populated")

	runs, results, err := is.ListRuns(ctx, opts.DestinationNamespace, inst.Name)
	require.NoError(t, err, "could not list runs in the destination database")
	assert.Len(t, runs, 5, "expected 5 runs") // dry-run, failed install, successful install, upgrade, uninstall

	lastRun := runs[4]
	assert.Equal(t, storage.InstallationSchemaVersion, lastRun.SchemaVersion, "incorrect run schema version")
	assert.Equal(t, "01G1VJQJV0RN5AW5BSZHNTVYTV", lastRun.ID, "incorrect run id")
	assert.Equal(t, "01G1VJQJV0RN5AW5BSZNJ1G6R7", lastRun.Revision, "incorrect run revision")
	assert.Equal(t, inst.Namespace, lastRun.Namespace, "incorrect run namespace")
	assert.Equal(t, inst.Name, lastRun.Installation, "incorrect run installation name")
	assert.Empty(t, lastRun.BundleReference, "We didn't track bundle reference in v0, so this can't be populated")
	assert.Empty(t, lastRun.BundleDigest, "We didn't track bundle digest in v0, so this can't be populated")
	assert.Equal(t, "uninstall", lastRun.Action, "incorrect run action")
	assert.Empty(t, lastRun.Custom, "We didn't set custom datadata on claims in v0, so this can't be populated")
	assert.Equal(t, "2022-04-29T16:13:20.48026-05:00", lastRun.Created.Format(time.RFC3339Nano), "incorrect run created timestamp")
	assert.Empty(t, lastRun.ParameterSets, "We didn't track run parameter sets in v0, so this can't be populated")
	assert.Empty(t, lastRun.ParameterOverrides, "We didn't track run parameter overrides in v0, so this can't be populated")
	assert.Empty(t, lastRun.CredentialSets, "We didn't track run credential sets in v0, so this can't be populated")
	assert.Len(t, lastRun.Parameters.Parameters, 1, "expected one parameter set on the run")
	params := lastRun.Parameters.Parameters
	assert.Equal(t, "porter-debug", params[0].Name, "expected the porter-debug parameter to be set on the run")
	assert.Equal(t, "value", params[0].Source.Key, "expected the porter-debug parameter to be a hard-coded value")
	assert.Equal(t, "true", params[0].Source.Value, "expected the porter-debug parameter to be false")

	runResults := results[lastRun.ID]
	assert.Len(t, runResults, 2, "expected 2 results for the last run")

	lastResult := runResults[1]
	assert.Equal(t, "01G1VJQM4AVN7SCXC8WV3M0D7N", lastResult.ID, "incorrect result id")
	assert.Equal(t, lastRun.ID, lastResult.RunID, "incorrect result id")
	assert.Equal(t, inst.Name, lastResult.Installation, "incorrect result installation name")
	assert.Equal(t, inst.Namespace, lastResult.Namespace, "incorrect result namespace")
	assert.Equal(t, "succeeded", lastResult.Status, "incorrect result status")
	assert.Equal(t, "yipee", lastResult.Message, "incorrect result message")
	assert.Empty(t, lastResult.Custom, "We didn't track custom metadata on results in v0, so this can't be populated")
	assert.Equal(t, "2022-04-29T16:13:21.802457-05:00", lastResult.Created.Format(time.RFC3339Nano), "invalid result created timestamp")
	assert.Contains(t, lastResult.OutputMetadata, cnab.OutputInvocationImageLogs, "expected the logs to be captured as an output")
	digest, _ := lastResult.OutputMetadata.GetContentDigest(cnab.OutputInvocationImageLogs)
	assert.Equal(t, "sha256:a7fdc86f826691ae1c6bf6bbcba43abbb67cf45b45093652a98327521a62d69c", digest, "incorrect content digest for logs output")
	generatedFlag, ok := lastResult.OutputMetadata.GetGeneratedByBundle(cnab.OutputInvocationImageLogs)
	assert.True(t, ok, "expected a generated flag to be set on the logs output")
	assert.False(t, generatedFlag, "incorrect content digest for logs output")

	logsOutput, err := is.GetLastOutput(ctx, inst.Namespace, inst.Name, cnab.OutputInvocationImageLogs)
	require.NoError(t, err, "error retrieving the last set of logs for the installation")
	test.CompareGoldenFile(t, "testdata/v0_home/outputs/01G1VJQM4AVN7SCXC8WV3M0D7N/01G1VJQM4AVN7SCXC8WV3M0D7N-io.cnab.outputs.invocationImageLogs", string(logsOutput.Value))
}
