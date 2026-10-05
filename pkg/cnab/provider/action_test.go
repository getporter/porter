package cnabprovider

import (
	"context"
	"encoding/json"
	"errors"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/pkg/test"
	"os"
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// failingSecretsStore wraps a real secrets.Store, forcing Create to fail, to
// simulate a secret store outage when persisting a sensitive output.
type failingSecretsStore struct {
	secrets.Store
	createErr error
}

func (f failingSecretsStore) Create(ctx context.Context, keyName, keyValue, value string) error {
	if f.createErr != nil {
		return f.createErr
	}
	return f.Store.Create(ctx, keyName, keyValue, value)
}

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

	err := d.SaveOperationResult(ctx, opResult, i, run, result, false)
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

	err := d.SaveOperationResult(ctx, opResult, i, run, result, false)
	require.NoError(t, err)

	outputs, err := d.TestInstallations.GetOutputs(ctx, run.ID)
	require.NoError(t, err)

	_, hasPorterState := outputs.GetByName("porter-state")
	assert.True(t, hasPorterState, "porter-state should be saved for modifies:true actions")
}

func TestSaveOperationResult_SensitiveOutputPersistFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tc := config.NewTestConfig(t)
	testStorage := storage.NewTestStore(tc)
	testSecrets := secrets.NewTestSecretsProvider()
	testInstallations := storage.NewTestInstallationProviderFor(tc.TestContext.T, testStorage)
	testCredentials := storage.NewTestCredentialProviderFor(tc.TestContext.T, testStorage, testSecrets)
	testParameters := storage.NewTestParameterProviderFor(tc.TestContext.T, testStorage, testSecrets)

	createErr := errors.New("secret store unreachable")
	failingSecrets := failingSecretsStore{Store: testSecrets, createErr: createErr}

	d := NewTestRuntimeFor(tc, testInstallations, testCredentials, testParameters, failingSecrets)
	defer d.Close()

	instName := "mybuns"
	sensitive := true
	bun := bundle.Bundle{
		Actions: map[string]bundle.Action{
			"install": {Modifies: true},
		},
		Outputs: map[string]bundle.Output{
			"secret-output": {Definition: "secret-output", Path: "/cnab/app/outputs/secret-output"},
		},
		Definitions: map[string]*definition.Schema{
			"secret-output": {Type: "string", WriteOnly: &sensitive},
		},
	}
	i := d.TestInstallations.CreateInstallation(storage.NewInstallation("", instName), d.TestInstallations.SetMutableInstallationValues)
	run := storage.NewRun("", instName)
	run.Bundle = bun
	run.Action = "install"
	run = d.TestInstallations.CreateRun(run, d.TestInstallations.SetMutableRunValues)

	opResult := driver.OperationResult{
		Outputs: map[string]string{
			"secret-output": "top-secret-value",
		},
	}

	t.Run("default does not fail the command", func(t *testing.T) {
		result := run.NewResult(cnab.StatusSucceeded)
		err := d.SaveOperationResult(ctx, opResult, i, run, result, false)
		require.NoError(t, err, "a secret persist failure should not fail the command by default")

		outputs, err := d.TestInstallations.GetOutputs(ctx, run.ID)
		require.NoError(t, err)
		output, ok := outputs.GetByName("secret-output")
		require.True(t, ok)
		assert.Empty(t, output.Key, "no dangling key reference should be persisted")
		assert.Empty(t, output.Value, "the raw sensitive value should never be persisted to the primary store")
		assert.NotEmpty(t, output.PersistError)

		savedResult, err := d.TestInstallations.GetResult(ctx, result.ID)
		require.NoError(t, err)
		assert.True(t, savedResult.OutputPersistFailed)

		savedInstallation, err := d.TestInstallations.GetInstallation(ctx, i.Namespace, i.Name)
		require.NoError(t, err)
		assert.True(t, savedInstallation.Status.OutputPersistFailed)
		assert.Equal(t, cnab.StatusSucceeded, savedInstallation.Status.ResultStatus, "the installation itself should still show as succeeded")
	})

	t.Run("opt-in flag fails the command", func(t *testing.T) {
		result := run.NewResult(cnab.StatusSucceeded)
		err := d.SaveOperationResult(ctx, opResult, i, run, result, true)
		require.Error(t, err, "a secret persist failure should fail the command when opted in")
	})
}

func TestCheckForActiveRun(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d := NewTestRuntime(t)
	defer d.Close()

	instName := "mybuns"
	i := d.TestInstallations.CreateInstallation(storage.NewInstallation("", instName), d.TestInstallations.SetMutableInstallationValues)

	t.Run("no runs yet", func(t *testing.T) {
		err := d.checkForActiveRun(ctx, i, "")
		require.NoError(t, err)
	})

	blockingRun := d.TestInstallations.CreateRun(storage.NewRun("", instName), d.TestInstallations.SetMutableRunValues)
	d.TestInstallations.CreateResult(blockingRun.NewResult(cnab.StatusRunning))

	t.Run("an incomplete run blocks", func(t *testing.T) {
		err := d.checkForActiveRun(ctx, i, "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInstallationLocked{}))
	})

	t.Run("the run itself is excluded from the check", func(t *testing.T) {
		err := d.checkForActiveRun(ctx, i, blockingRun.ID)
		require.NoError(t, err)
	})

	d.TestInstallations.CreateResult(blockingRun.NewResult(cnab.StatusSucceeded))

	t.Run("a completed run does not block", func(t *testing.T) {
		err := d.checkForActiveRun(ctx, i, "")
		require.NoError(t, err)
	})
}

func TestExecute_InstallationLocked(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d := NewTestRuntime(t)
	defer d.Close()

	instName := "mybuns"
	bun := d.LoadTestBundle("testdata/bundle.json")

	i := d.TestInstallations.CreateInstallation(storage.NewInstallation("", instName), d.TestInstallations.SetMutableInstallationValues)

	blockingRun := d.TestInstallations.CreateRun(storage.NewRun("", instName), d.TestInstallations.SetMutableRunValues)
	d.TestInstallations.CreateResult(blockingRun.NewResult(cnab.StatusRunning))

	newArgs := func() ActionArguments {
		run := i.NewRun("zombies", bun)
		run.Bundle = bun.Bundle
		return ActionArguments{
			Run:             run,
			Installation:    i,
			BundleReference: cnab.BundleReference{Definition: bun},
		}
	}

	t.Run("flag disabled: proceeds despite the active run", func(t *testing.T) {
		err := d.Execute(ctx, newArgs())
		require.NoError(t, err)
	})

	t.Run("flag enabled: blocked by the active run", func(t *testing.T) {
		d.TestConfig.SetExperimentalFlags(experimental.FlagDependenciesV2)
		defer d.TestConfig.SetExperimentalFlags(0)

		err := d.Execute(ctx, newArgs())
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInstallationLocked{}))
	})

	t.Run("flag enabled and force-run: bypasses the lock", func(t *testing.T) {
		d.TestConfig.SetExperimentalFlags(experimental.FlagDependenciesV2)
		defer d.TestConfig.SetExperimentalFlags(0)

		args := newArgs()
		args.ForceRun = true
		err := d.Execute(ctx, args)
		require.NoError(t, err)
	})
}

func TestAddEnvironment(t *testing.T) {
	t.Parallel()

	args := ActionArguments{
		Installation: storage.Installation{
			ID: "myid",
			InstallationSpec: storage.InstallationSpec{
				Namespace: "myns",
				Name:      "mybuns",
			},
		},
	}

	// Make a context with a span that can be passed into the bundle
	provider := sdktrace.NewTracerProvider()
	spanCtx, span := provider.Tracer("").Start(context.Background(), t.Name())
	t.Cleanup(func() {
		span.End()
		require.NoError(t, provider.Shutdown(context.Background()))
	})

	t.Run("telemetry disabled", func(t *testing.T) {
		t.Parallel()

		d := NewTestRuntime(t)
		defer d.Close()
		d.Data.Verbosity = "warn"

		op := &driver.Operation{}
		err := d.AddEnvironment(spanCtx, args)(op)
		require.NoError(t, err, "AddEnvironment failed")

		want := map[string]string{
			config.EnvPorterInstallationNamespace: "myns",
			config.EnvPorterInstallationName:      "mybuns",
			config.EnvPorterInstallationID:        "myid",
			config.EnvPorterVerbosity:             "warn",
			portercontext.EnvCorrelationID:        d.CorrelationID(),
		}
		assert.Equal(t, want, op.Environment)
		assert.NotEmpty(t, d.CorrelationID())
	})

	t.Run("telemetry enabled", func(t *testing.T) {
		t.Parallel()

		d := NewTestRuntime(t)
		defer d.Close()
		d.Data.Telemetry = config.TelemetryConfig{
			Enabled:        true,
			Endpoint:       "collector:4317",
			Protocol:       "grpc",
			Insecure:       true,
			Certificate:    "/home/me/cert.pem",
			Headers:        map[string]string{"token": "secret"},
			Timeout:        "3s",
			RedirectToFile: true,
		}

		op := &driver.Operation{Environment: map[string]string{}}
		err := d.AddEnvironment(spanCtx, args)(op)
		require.NoError(t, err, "AddEnvironment failed")

		want := map[string]string{
			config.EnvPorterInstallationNamespace: "myns",
			config.EnvPorterInstallationName:      "mybuns",
			config.EnvPorterInstallationID:        "myid",
			config.EnvPorterVerbosity:             "debug",
			portercontext.EnvCorrelationID:        d.CorrelationID(),
			"PORTER_TELEMETRY_ENABLED":            "true",
			"PORTER_TELEMETRY_ENDPOINT":           "collector:4317",
			"PORTER_TELEMETRY_PROTOCOL":           "grpc",
			"PORTER_TELEMETRY_INSECURE":           "true",
			"PORTER_TELEMETRY_TIMEOUT":            "3s",
			"TRACEPARENT":                         portercontext.TraceEnviron(spanCtx)["TRACEPARENT"],
		}
		assert.Equal(t, want, op.Environment, "only the safe subset of the telemetry settings should be passed into the bundle")
		assert.Contains(t, op.Environment["TRACEPARENT"], span.SpanContext().TraceID().String())
	})

	t.Run("telemetry enabled, no span", func(t *testing.T) {
		t.Parallel()

		d := NewTestRuntime(t)
		defer d.Close()
		d.Data.Telemetry.Enabled = true

		op := &driver.Operation{}
		err := d.AddEnvironment(context.Background(), args)(op)
		require.NoError(t, err, "AddEnvironment failed")

		assert.Equal(t, "true", op.Environment["PORTER_TELEMETRY_ENABLED"])
		assert.Equal(t, "false", op.Environment["PORTER_TELEMETRY_INSECURE"], "insecure should be explicitly disabled so the bundle image can't override it")
		assert.NotContains(t, op.Environment, "TRACEPARENT")
	})
}

// A bundle image can define its own telemetry settings. Validate that when the
// host requires TLS, the bundle does too, even when the image says otherwise.
func TestAddEnvironment_HostRequiresTLS_OverridesBundleImage(t *testing.T) {
	// Do not run in parallel since we use t.Setenv

	d := NewTestRuntime(t)
	defer d.Close()
	d.Data.Telemetry = config.TelemetryConfig{
		Enabled:  true,
		Endpoint: "collector:4317",
		Insecure: false,
	}

	op := &driver.Operation{}
	err := d.AddEnvironment(context.Background(), ActionArguments{})(op)
	require.NoError(t, err, "AddEnvironment failed")

	// The environment inside the bundle is what was defined in the image,
	// overridden by what porter passes in when it runs the bundle
	bundleEnv := map[string]string{
		"PORTER_TELEMETRY_INSECURE":   "true",
		"OTEL_EXPORTER_OTLP_INSECURE": "true",
	}
	for k, v := range op.Environment {
		bundleEnv[k] = v
	}
	for k, v := range bundleEnv {
		t.Setenv(k, v)
	}

	// Load the configuration like the porter runtime and mixins do inside the bundle
	bundleCfg := config.NewTestConfig(t)
	bundleCfg.DataLoader = config.LoadFromEnvironment()
	_, err = bundleCfg.Load(context.Background(), nil)
	require.NoError(t, err, "Load failed")

	assert.True(t, bundleCfg.Data.Telemetry.Enabled, "telemetry should be enabled inside the bundle")
	assert.Equal(t, "collector:4317", bundleCfg.Data.Telemetry.Endpoint)
	assert.False(t, bundleCfg.Data.Telemetry.Insecure, "the bundle image should not be able to turn off TLS when the host requires it")
}
