package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/secrets/host"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	kahnlatest = cnab.MustParseOCIReference("deislabs/kubekahn:latest")
)

func TestInstallFromTagIgnoresCurrentBundle(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	err := p.Create()
	require.NoError(t, err)

	installOpts := NewInstallOptions()
	installOpts.Reference = "mybun:1.0"

	err = installOpts.Validate(context.Background(), []string{}, p.Porter)
	require.NoError(t, err)

	assert.Empty(t, installOpts.File, "The install should ignore the bundle in the current directory because we are installing from a tag")
}

func TestPorter_BuildActionArgs(t *testing.T) {
	ctx := context.Background()

	t.Run("no bundle set", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()
		opts := NewInstallOptions()
		opts.Name = "mybuns"

		err := opts.Validate(ctx, nil, p.Porter)
		require.Error(t, err, "Validate should fail")
		assert.Contains(t, err.Error(), "No bundle specified")
	})

	t.Run("porter.yaml set", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		opts := NewInstallOptions()
		opts.File = "porter.yaml"
		p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
		p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", ".cnab/bundle.json")

		err := opts.Validate(ctx, nil, p.Porter)
		require.NoError(t, err, "Validate failed")
		args, err := p.BuildActionArgs(ctx, storage.Installation{}, opts)
		require.NoError(t, err, "BuildActionArgs failed")

		assert.NotEmpty(t, args.BundleReference.Definition)
	})

	// Just do a quick check that things are populated correctly when a bundle.json is passed
	t.Run("bundle.json set", func(t *testing.T) {
		p := NewTestPorter(t)
		opts := NewInstallOptions()
		opts.CNABFile = "/bundle.json"
		p.TestConfig.TestContext.AddTestFile("testdata/bundle.json", "/bundle.json")

		err := opts.Validate(ctx, nil, p.Porter)
		require.NoError(t, err, "Validate failed")
		args, err := p.BuildActionArgs(ctx, storage.Installation{}, opts)
		require.NoError(t, err, "BuildActionArgs failed")

		assert.NotEmpty(t, args.BundleReference.Definition, "BundlePath was not populated correctly")
	})

	t.Run("remaining fields", func(t *testing.T) {
		p := NewTestPorter(t)
		p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
		p.TestConfig.TestContext.AddTestFileFromRoot("pkg/runtime/testdata/relocation-mapping.json", "relocation-mapping.json")
		opts := InstallOptions{
			BundleExecutionOptions: &BundleExecutionOptions{
				AllowDockerHostAccess: true,
				DebugMode:             true,
				Params: []string{
					"my-first-param=1",
				},
				ParameterSets: []string{
					"porter-hello",
				},
				CredentialIdentifiers: []string{
					"mycreds",
				},
				Driver: "docker",
				BundleReferenceOptions: &BundleReferenceOptions{
					installationOptions: installationOptions{
						bundleFileOptions: bundleFileOptions{
							RelocationMapping: "relocation-mapping.json",
							File:              config.Name,
						},
						Name: "MyInstallation",
					},
				},
			},
		}
		p.TestParameters.AddSecret("PARAM2_SECRET", "VALUE2")
		p.TestParameters.AddTestParameters("testdata/paramset2.json")

		err := opts.Validate(ctx, nil, p.Porter)
		require.NoError(t, err, "Validate failed")
		existingInstall := storage.Installation{Name: opts.Name}
		args, err := p.BuildActionArgs(ctx, existingInstall, opts)
		require.NoError(t, err, "BuildActionArgs failed")

		expectedParams := map[string]interface{}{
			"my-first-param":  1,
			"my-second-param": "VALUE2",
			"porter-debug":    true,
			"porter-state":    nil,
		}

		assert.Equal(t, opts.AllowDockerHostAccess, args.AllowDockerHostAccess, "AllowDockerHostAccess not populated correctly")
		assert.Equal(t, opts.Driver, args.Driver, "Driver not populated correctly")
		assert.EqualValues(t, expectedParams, args.Params, "Params not populated correctly")
		assert.Equal(t, existingInstall, args.Installation, "Installation not populated correctly")
		wantReloMap := relocation.ImageRelocationMap{"gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687": "my.registry/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687"}
		assert.Equal(t, wantReloMap, args.BundleReference.RelocationMap, "RelocationMapping not populated correctly")
	})
}

func TestManifestIgnoredWithTag(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	t.Run("ignore manifest in cwd if tag present", func(t *testing.T) {
		opts := BundleReferenceOptions{}
		opts.Reference = "deislabs/kubekahn:latest"

		// `path.Join(wd...` -> makes cnab.go#defaultBundleFiles#manifestExists `true`
		// Only when `manifestExists` eq to `true`, default bundle logic will run
		p.TestConfig.TestContext.AddTestFileContents([]byte(""), config.Name)
		// When execution reach to `readFromFile`, manifest file path will be lost.
		// So, had to use root manifest file also for error simuation purpose
		p.TestConfig.TestContext.AddTestFileContents([]byte(""), config.Name)

		err := opts.Validate(context.Background(), nil, p.Porter)
		require.NoError(t, err, "Validate failed")
	})
}

func TestBundleActionOptions_Validate(t *testing.T) {
	t.Run("driver flag unset", func(t *testing.T) {
		p := NewTestPorter(t)
		p.DataLoader = config.LoadFromEnvironment()
		p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", []byte("runtime-driver: kubernetes"), pkg.FileModeWritable)
		require.NoError(t, p.Connect(context.Background()))

		opts := NewInstallOptions()
		opts.Reference = "ghcr.io/getporter/examples/porter-hello:v0.2.0"
		require.NoError(t, opts.Validate(context.Background(), nil, p.Porter))
		assert.Equal(t, "kubernetes", opts.Driver)
	})
	t.Run("driver flag set", func(t *testing.T) {
		p := NewTestPorter(t)
		p.DataLoader = config.LoadFromEnvironment()
		p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", []byte("driver: kubernetes"), pkg.FileModeWritable)
		require.NoError(t, p.Connect(context.Background()))

		opts := NewInstallOptions()
		opts.Driver = "docker"
		opts.Reference = "ghcr.io/getporter/examples/porter-hello:v0.2.0"
		require.NoError(t, opts.Validate(context.Background(), nil, p.Porter))
		assert.Equal(t, "docker", opts.Driver)
	})
}

func TestBundleExecutionOptions_defaultDriver(t *testing.T) {
	t.Run("no driver specified", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		opts := NewBundleExecutionOptions()

		opts.defaultDriver(p.Porter)

		assert.Equal(t, "docker", opts.Driver, "expected the driver value to default to docker")
	})

	t.Run("driver flag set", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()

		opts := NewBundleExecutionOptions()
		opts.Driver = "kubernetes"

		opts.defaultDriver(p.Porter)

		assert.Equal(t, "kubernetes", opts.Driver, "expected the --driver flag value to be used")
	})

	t.Run("allow docker host access defaults to config", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()
		p.Config.Data.AllowDockerHostAccess = true

		opts := NewBundleExecutionOptions()

		opts.defaultDriver(p.Porter)

		assert.True(t, opts.AllowDockerHostAccess, "expected allow-docker-host-access to inherit the value from the config file when the flag isn't specified")
	})

	t.Run("allow docker host access flag set", func(t *testing.T) {
		p := NewTestPorter(t)
		defer p.Close()
		p.Config.Data.AllowDockerHostAccess = false

		opts := NewBundleExecutionOptions()
		opts.AllowDockerHostAccess = true

		opts.defaultDriver(p.Porter)

		assert.True(t, opts.AllowDockerHostAccess, "expected allow-docker-host-access to use the flag value when specified")
	})

}

func TestBundleExecutionOptions_ParseParamSets(t *testing.T) {
	ctx := context.Background()
	p := NewTestPorter(t)
	defer p.Close()

	p.AddTestFile("testdata/porter.yaml", "porter.yaml")
	p.TestParameters.AddSecret("foo_secret", "foo_value")
	p.TestParameters.AddSecret("PARAM2_SECRET", "VALUE2")
	p.TestParameters.AddTestParameters("testdata/paramset2.json")

	opts := NewBundleExecutionOptions()
	opts.ParameterSets = []string{"porter-hello"}

	err := opts.Validate(ctx, []string{}, p.Porter)
	assert.NoError(t, err)

	err = opts.parseParamSets(ctx, p.Porter, cnab.ExtendedBundle{})
	assert.NoError(t, err)

	wantParams := map[string]string{
		"my-second-param": "VALUE2",
	}
	assert.Equal(t, wantParams, opts.parsedParamSets, "resolved unexpected parameter values")
}

func TestBundleExecutionOptions_ParseParamSets_Failed(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/porter-with-file-param.yaml", config.Name)
	p.TestConfig.TestContext.AddTestFile("testdata/paramset-with-file-param.json", "/paramset.json")

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, p.Config, config.Name)
	require.NoError(t, err)
	bun, err := configadapter.ConvertToTestBundle(ctx, p.Config, m)
	require.NoError(t, err)

	opts := NewBundleExecutionOptions()
	opts.ParameterSets = []string{
		"/paramset.json",
	}

	err = opts.Validate(ctx, []string{}, p.Porter)
	assert.NoError(t, err)

	err = opts.parseParamSets(ctx, p.Porter, bun)
	assert.Error(t, err)

}

func TestBundleExecutionOptions_LoadParameters(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", config.Name)

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, p.Config, config.Name)
	require.NoError(t, err)
	bun, err := configadapter.ConvertToTestBundle(ctx, p.Config, m)
	require.NoError(t, err)

	opts := NewBundleExecutionOptions()
	opts.Params = []string{"my-first-param=1", "my-second-param=2"}

	err = opts.LoadParameters(context.Background(), p.Porter, bun)
	require.NoError(t, err)

	assert.Len(t, opts.Params, 2)
}

func TestBundleExecutionOptions_CombineParameters(t *testing.T) {
	c := portercontext.NewTestContext(t)

	t.Run("no override present, no parameter set present", func(t *testing.T) {
		opts := NewBundleExecutionOptions()

		params := opts.combineParameters(c.Context)
		require.Equal(t, map[string]string{}, params,
			"expected combined params to be empty")
	})

	t.Run("override present, no parameter set present", func(t *testing.T) {
		opts := NewBundleExecutionOptions()
		opts.parsedParams = map[string]string{
			"foo": "foo_cli_override",
		}

		params := opts.combineParameters(c.Context)
		require.Equal(t, "foo_cli_override", params["foo"],
			"expected param 'foo' to have override value")
	})

	t.Run("no override present, parameter set present", func(t *testing.T) {
		opts := NewBundleExecutionOptions()
		opts.parsedParamSets = map[string]string{
			"foo": "foo_via_paramset",
		}

		params := opts.combineParameters(c.Context)
		require.Equal(t, "foo_via_paramset", params["foo"],
			"expected param 'foo' to have parameter set value")
	})

	t.Run("override present, parameter set present", func(t *testing.T) {
		opts := NewBundleExecutionOptions()
		opts.parsedParams = map[string]string{
			"foo": "foo_cli_override",
		}
		opts.parsedParamSets = map[string]string{
			"foo": "foo_via_paramset",
		}

		params := opts.combineParameters(c.Context)
		require.Equal(t, "foo_cli_override", params["foo"],
			"expected param 'foo' to have override value, which has precedence over the parameter set value")
	})

	t.Run("debug mode on", func(t *testing.T) {
		var opts BundleExecutionOptions
		opts.DebugMode = true
		debugContext := portercontext.NewTestContext(t)
		params := opts.combineParameters(debugContext.Context)
		require.Equal(t, "true", params["porter-debug"], "porter-debug should be set to true when p.Debug is true")
	})
}

func TestBundleExecutionOptions_populateInternalParameterSet(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	ctx := context.Background()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", config.Name)
	m, err := manifest.LoadManifestFrom(context.Background(), p.Config, config.Name)
	require.NoError(t, err)
	bun, err := configadapter.ConvertToTestBundle(ctx, p.Config, m)
	require.NoError(t, err)

	sensitiveParamName := "my-second-param"
	sensitiveParamValue := "2"
	nonsensitiveParamName := "my-first-param"
	nonsensitiveParamValue := "1"
	opts := NewBundleExecutionOptions()
	opts.Params = []string{nonsensitiveParamName + "=" + nonsensitiveParamValue, sensitiveParamName + "=" + sensitiveParamValue}

	err = opts.LoadParameters(ctx, p.Porter, bun)
	require.NoError(t, err)

	i := storage.NewInstallation("", bun.Name)

	err = opts.populateInternalParameterSet(ctx, p.Porter, bun, &i)
	require.NoError(t, err)

	require.Len(t, i.Parameters.Parameters, 2)

	// there should be no sensitive value on installation record
	for _, param := range i.Parameters.Parameters {
		if param.Name == sensitiveParamName {
			require.Equal(t, param.Source.Key, secrets.SourceSecret)
			require.NotEqual(t, param.Source.Value, sensitiveParamValue)
			continue
		}
		require.Equal(t, param.Source.Key, host.SourceValue)
		require.Equal(t, param.Source.Value, nonsensitiveParamValue)
	}

	// if no parameter override specified, installation record should be updated
	// as well
	opts.combinedParameters = nil
	opts.Params = make([]string, 0)
	err = opts.LoadParameters(ctx, p.Porter, bun)
	require.NoError(t, err)
	err = opts.populateInternalParameterSet(ctx, p.Porter, bun, &i)
	require.NoError(t, err)

	require.Len(t, i.Parameters.Parameters, 0)
}
