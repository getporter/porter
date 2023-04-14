package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	configadapter "get.porter.sh/porter/pkg/cnab/config-adapter"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"get.porter.sh/porter/tests"
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

		// pretend that we've resolved the parameters
		opts.finalParams = map[string]interface{}{}

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

		// pretend that we've resolved the parameters
		opts.finalParams = map[string]interface{}{}

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
						BundleDefinitionOptions: BundleDefinitionOptions{
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
		existingInstall := storage.NewInstallation(opts.Namespace, opts.Name)

		// resolve the parameters before building the action options to use for running the bundle
		err = p.applyActionOptionsToInstallation(ctx, opts, &existingInstall)
		require.NoError(t, err)

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
		assert.NotEmpty(t, args.Installation, "Installation not populated")
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
		ctx, err := p.Connect(context.Background())
		require.NoError(t, err)

		opts := NewInstallOptions()
		opts.Reference = "ghcr.io/getporter/examples/porter-hello:v0.2.0"
		require.NoError(t, opts.Validate(ctx, nil, p.Porter))
		assert.Equal(t, "kubernetes", opts.Driver)
	})
	t.Run("driver flag set", func(t *testing.T) {
		p := NewTestPorter(t)
		p.DataLoader = config.LoadFromEnvironment()
		p.FileSystem.WriteFile("/home/myuser/.porter/config.yaml", []byte("driver: kubernetes"), pkg.FileModeWritable)
		ctx, err := p.Connect(context.Background())
		require.NoError(t, err)

		opts := NewInstallOptions()
		opts.Driver = "docker"
		opts.Reference = "ghcr.io/getporter/examples/porter-hello:v0.2.0"
		require.NoError(t, opts.Validate(ctx, nil, p.Porter))
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
	p := NewTestPorter(t)
	defer p.Close()

	p.AddTestFile("testdata/porter.yaml", "porter.yaml")
	p.TestParameters.AddSecret("foo_secret", "foo_value")
	p.TestParameters.AddSecret("PARAM2_SECRET", "VALUE2")
	p.TestParameters.AddTestParameters("testdata/paramset2.json")

	ctx := context.Background()
	m, err := manifest.LoadManifestFrom(ctx, p.Config, config.Name)
	require.NoError(t, err)
	bun, err := configadapter.ConvertToTestBundle(ctx, p.Config, m)
	require.NoError(t, err)

	opts := NewUpgradeOptions()
	opts.ParameterSets = []string{"porter-hello"}
	opts.bundleRef = &cnab.BundleReference{Definition: bun}

	err = opts.Validate(ctx, []string{}, p.Porter)
	assert.NoError(t, err)

	inst := storage.NewInstallation("", "mybuns")
	err = p.applyActionOptionsToInstallation(ctx, opts, &inst)
	require.NoError(t, err)

	wantParams := map[string]interface{}{
		"my-second-param": "VALUE2",
		"porter-debug":    false,
		"porter-state":    nil,
	}
	assert.Equal(t, wantParams, opts.GetParameters(), "resolved unexpected parameter values")
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

	opts := NewInstallOptions()
	opts.ParameterSets = []string{
		"/paramset.json",
	}
	opts.bundleRef = &cnab.BundleReference{Definition: bun}

	err = opts.Validate(ctx, []string{}, p.Porter)
	assert.NoError(t, err)

	inst := storage.NewInstallation("myns", "mybuns")

	err = p.applyActionOptionsToInstallation(ctx, opts, &inst)
	tests.RequireErrorContains(t, err, "/paramset.json not found", "Porter no longer supports passing a parameter set file to the -p flag, validate that passing a file doesn't work")
}

// Validate that when an installation is run with a mix of overrides and parameter sets
// that it follows the rules for the paramter hierarchy
// highest -> lowest preceence
// - user override
// - previous value from last run
// - value resolved from a named parameter set
// - default value of the parameter
func TestPorter_applyActionOptionsToInstallation_FollowsParameterHierarchy(t *testing.T) {
	t.Parallel()

	p := NewTestPorter(t)
	defer p.Close()
	ctx := context.Background()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", config.Name)
	m, err := manifest.LoadManifestFrom(context.Background(), p.Config, config.Name)
	require.NoError(t, err)
	bun, err := configadapter.ConvertToTestBundle(ctx, p.Config, m)
	require.NoError(t, err)

	err = p.TestParameters.InsertParameterSet(ctx, storage.NewParameterSet("", "myps",
		storage.ValueStrategy("my-second-param", "via_paramset")))
	require.NoError(t, err, "Create my-second-param parameter set failed")

	makeOpts := func() InstallOptions {
		opts := NewInstallOptions()
		opts.BundleReferenceOptions.bundleRef = &cnab.BundleReference{
			Reference:  kahnlatest,
			Definition: bun,
		}
		return opts
	}

	t.Run("no override present, no parameter set present", func(t *testing.T) {
		i := storage.NewInstallation("", bun.Name)
		opts := makeOpts()
		err = p.applyActionOptionsToInstallation(ctx, opts, &i)
		require.NoError(t, err)

		finalParams := opts.GetParameters()
		wantParams := map[string]interface{}{
			"my-first-param":  9,
			"my-second-param": "spring-music-demo",
			"porter-debug":    false,
			"porter-state":    nil,
		}
		assert.Equal(t, wantParams, finalParams,
			"expected combined params to have the default parameter values from the bundle")
	})

	t.Run("override present, no parameter set present", func(t *testing.T) {
		i := storage.NewInstallation("", bun.Name)
		opts := makeOpts()
		opts.Params = []string{"my-second-param=cli_override"}
		err = p.applyActionOptionsToInstallation(ctx, opts, &i)
		require.NoError(t, err)

		finalParams := opts.GetParameters()
		require.Contains(t, finalParams, "my-second-param",
			"expected my-second-param to be a parameter")
		require.Equal(t, "cli_override", finalParams["my-second-param"],
			"expected param 'my-second-param' to be set with the override specified by the user")
	})

	t.Run("no override present, parameter set present", func(t *testing.T) {
		i := storage.NewInstallation("", bun.Name)
		opts := makeOpts()
		opts.ParameterSets = []string{"myps"}
		err = p.applyActionOptionsToInstallation(ctx, opts, &i)
		require.NoError(t, err)

		finalParams := opts.GetParameters()
		require.Contains(t, finalParams, "my-second-param",
			"expected my-second-param to be a parameter")
		require.Equal(t, finalParams["my-second-param"], "via_paramset",
			"expected param 'my-second-param' to be set with the value from the parameter set")
	})

	t.Run("override present, parameter set present", func(t *testing.T) {
		i := storage.NewInstallation("", bun.Name)
		opts := makeOpts()
		opts.Params = []string{"my-second-param=cli_override"}
		opts.ParameterSets = []string{"myps"}
		err = p.applyActionOptionsToInstallation(ctx, opts, &i)
		require.NoError(t, err)

		finalParams := opts.GetParameters()
		require.Contains(t, finalParams, "my-second-param",
			"expected my-second-param to be a parameter")
		require.Equal(t, finalParams["my-second-param"], "cli_override",
			"expected param 'my-second-param' to be set with the value of the user override, which has precedence over the parameter set value")
	})

	t.Run("debug mode on", func(t *testing.T) {
		i := storage.NewInstallation("", bun.Name)
		opts := makeOpts()
		opts.DebugMode = true
		err = p.applyActionOptionsToInstallation(ctx, opts, &i)
		require.NoError(t, err)

		finalParams := opts.GetParameters()
		debugParam, ok := finalParams["porter-debug"]
		require.True(t, ok, "expected porter-debug to be set")
		require.Equal(t, true, debugParam, "expected porter-debug to be true")
	})
}

// Validate that when we resolve parameters on an installation that sensitive parameters are
// not persisted on the installation record and instead are referenced by a secret
func TestPorter_applyActionOptionsToInstallation_sanitizesParameters(t *testing.T) {
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
	opts := NewInstallOptions()
	opts.BundleReferenceOptions.bundleRef = &cnab.BundleReference{
		Reference:  kahnlatest,
		Definition: bun,
	}
	opts.Params = []string{nonsensitiveParamName + "=" + nonsensitiveParamValue, sensitiveParamName + "=" + sensitiveParamValue}

	i := storage.NewInstallation("", bun.Name)

	err = p.applyActionOptionsToInstallation(ctx, opts, &i)
	require.NoError(t, err)
	require.Len(t, i.Parameters.Parameters, 2)

	// there should be no sensitive value on installation record
	for _, param := range i.Parameters.Parameters {
		if param.Name == sensitiveParamName {
			require.Equal(t, param.Source.Strategy, secrets.SourceSecret)
			require.NotEqual(t, param.Source.Hint, sensitiveParamValue)
			continue
		}
		require.Equal(t, param.Source.Strategy, host.SourceValue)
		require.Equal(t, param.Source.Hint, nonsensitiveParamValue)
	}

	// When no parameter override specified, installation record should be updated
	// as well
	opts = NewInstallOptions()
	opts.BundleReferenceOptions.bundleRef = &cnab.BundleReference{
		Reference:  kahnlatest,
		Definition: bun,
	}
	err = p.applyActionOptionsToInstallation(ctx, opts, &i)
	require.NoError(t, err)

	// Check that when no parameter overrides are specified, we use the originally specified parameters from the previous run
	require.Len(t, i.Parameters.Parameters, 2)
	require.Equal(t, "my-first-param", i.Parameters.Parameters[0].Name)
	require.Equal(t, "1", i.Parameters.Parameters[0].Source.Hint)
	require.Equal(t, "my-second-param", i.Parameters.Parameters[1].Name)
	require.Equal(t, "secret", i.Parameters.Parameters[1].Source.Strategy)
}

// When the installation has been used before with a parameter value
// the previous param value should be updated and the other previous values
// that were not set this time around should be re-used.
// i.e. you should be able to run
// porter install --param logLevel=debug --param featureA=enabled
// porter upgrade --param logLevel=info
// and when upgrade is run, the old value for featureA is kept
func TestPorter_applyActionOptionsToInstallation_PreservesExistingParams(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Close()

	ctx := context.Background()

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", config.Name)
	m, err := manifest.LoadManifestFrom(context.Background(), p.Config, config.Name)
	require.NoError(t, err)
	bun, err := configadapter.ConvertToTestBundle(ctx, p.Config, m)
	require.NoError(t, err)

	nonsensitiveParamName := "my-first-param"
	nonsensitiveParamValue := "3"
	opts := NewUpgradeOptions()
	opts.BundleReferenceOptions.bundleRef = &cnab.BundleReference{
		Reference:  kahnlatest,
		Definition: bun,
	}
	opts.Params = []string{nonsensitiveParamName + "=" + nonsensitiveParamValue}

	i := storage.NewInstallation("", bun.Name)
	i.Parameters = storage.NewParameterSet("", "internal-ps",
		storage.ValueStrategy("my-first-param", "1"),
		storage.ValueStrategy("my-second-param", "2"),
	)

	err = p.applyActionOptionsToInstallation(ctx, opts, &i)
	require.NoError(t, err)
	require.Len(t, i.Parameters.Parameters, 2)

	// Check that overrides are applied on top of existing parameters
	require.Len(t, i.Parameters.Parameters, 2)
	require.Equal(t, "my-first-param", i.Parameters.Parameters[0].Name)
	require.Equal(t, "value", i.Parameters.Parameters[0].Source.Strategy, "my-first-param isn't sensitive and can be stored in a hard-coded value")
	require.Equal(t, "my-second-param", i.Parameters.Parameters[1].Name)
	require.Equal(t, "secret", i.Parameters.Parameters[1].Source.Strategy, "my-second-param should be stored on the installation using a secret since it's sensitive")

	// Check the values stored are correct
	params, err := p.Parameters.ResolveAll(ctx, i.Parameters)
	require.NoError(t, err, "Failed to resolve the installation parameters")
	require.Equal(t, secrets.Set{
		"my-first-param":  "3", // Should have used the override
		"my-second-param": "2", // Should have kept the existing value from the last run
	}, params, "Incorrect parameter values were persisted on the installationß")
}
