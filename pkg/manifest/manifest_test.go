package manifest

import (
	"context"
	"os"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/schema"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadManifest(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	require.NotNil(t, m, "manifest was nil")
	require.Equal(t, m.Name, "hello", "manifest has incorrect name")
	require.Equal(t, m.Description, "An example Porter configuration", "manifest has incorrect description")
	require.Equal(t, m.Version, "0.1.0", "manifest has incorrect version")
	require.Equal(t, m.Registry, "localhost:5000", "manifest has incorrect registry")
	require.Equal(t, m.Reference, "localhost:5000/hello:v0.1.0", "manifest has incorrect reference")

	require.Len(t, m.Maintainers, 4, "manifest has incorrect number of maintainers")

	john, jane, janine, mike := m.Maintainers[0], m.Maintainers[1], m.Maintainers[2], m.Maintainers[3]
	require.Equal(t, "John Doe", john.Name, "manifest: Maintainer name is incorrect")
	require.Equal(t, "john.doe@example.com", john.Email, "manifest: Maintainer email is incorrect")
	require.Equal(t, "https://example.com/a", john.Url, "manifest: Maintainer url is incorrect")

	require.Equal(t, "Jane Doe", jane.Name, "manifest: Maintainer name is incorrect")
	require.Equal(t, "", jane.Email, "manifest: Maintainer email is incorrect")
	require.Equal(t, "https://example.com/b", jane.Url, "manifest: Maintainer url is incorrect")

	require.Equal(t, "Janine Doe", janine.Name, "manifest: Maintainer name is incorrect")
	require.Equal(t, "janine.doe@example.com", janine.Email, "manifest: Maintainer email is incorrect")
	require.Equal(t, "", janine.Url, "manifest: Maintainer url is incorrect")

	require.Equal(t, "", mike.Name, "manifest: Maintainer name is incorrect")
	require.Equal(t, "mike.doe@example.com", mike.Email, "manifest: Maintainer email is incorrect")
	require.Equal(t, "https://example.com/c", mike.Url, "manifest: Maintainer url is incorrect")

	assert.Equal(t, []MixinDeclaration{{Name: "exec"}}, m.Mixins, "expected manifest to declare the exec mixin")
	require.Len(t, m.Install, 1, "expected 1 install step")

	installStep := m.Install[0]
	description, _ := installStep.GetDescription()
	assert.NotNil(t, description, "expected the install description to be populated")

	mixin := installStep.GetMixinName()
	assert.Equal(t, "exec", mixin, "incorrect install step mixin used")

	require.Len(t, m.CustomActions, 1, "expected manifest to declare 1 custom action")
	require.Contains(t, m.CustomActions, "status", "expected manifest to declare a status action")

	statusStep := m.CustomActions["status"][0]
	description, _ = statusStep.GetDescription()
	assert.Equal(t, "Get World Status", description, "unexpected status step description")

	mixin = statusStep.GetMixinName()
	assert.Equal(t, "exec", mixin, "unexpected status step mixin")
}

func TestLoadManifestWithDependencies(t *testing.T) {
	// Make sure that we can parse the bundle in both v1 dep mode and v2 dep mode
	testcases := []struct {
		name          string
		depsv2enabled bool
	}{
		{"deps v1", false},
		{"deps v2", true},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			c := config.NewTestConfig(t)
			if tc.depsv2enabled {
				c.SetExperimentalFlags(experimental.FlagDependenciesV2)
			}

			c.TestContext.AddTestFile("testdata/porter.yaml", config.Name)
			c.TestContext.AddTestDirectory("testdata/bundles", "bundles")

			m, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
			require.NoError(t, err, "could not load manifest")

			require.NotNil(t, m)
			assert.Equal(t, []MixinDeclaration{{Name: "exec"}}, m.Mixins)
			require.Len(t, m.Install, 1)

			installStep := m.Install[0]
			description, _ := installStep.GetDescription()
			require.NotNil(t, description)

			mixin := installStep.GetMixinName()
			assert.Equal(t, "exec", mixin)

			require.Len(t, m.Dependencies.Requires, 1, "expected one dependency")
			dep := m.Dependencies.Requires[0]
			assert.Equal(t, "getporter/azure-mysql:5.7", dep.Bundle.Reference, "expected the dependency to be set")
			assert.Equal(t, "5.7.x", dep.Bundle.Version, "expected the version range to be set")
			assert.Equal(t, map[string]string{"database-name": "wordpress"}, dep.Parameters, "expected the dependency parameters to be set")

			// The remaining fields are only supported in depsv2 but the manifest still parses them. It's only a behavior difference if we act on the information or not.
			assert.Equal(t, map[string]string{"password": "mcstuffins"}, dep.Credentials, "expected the dependency credentials to be set")

			// TODO(PEP003) validate the bundle interface document
			/*
				wantDoc := &BundleInterfaceDocument{
					Parameters: map[string]ParameterDefinition{
						"password": {
							Name:   "password",
							Schema: definition.Schema{Type: "string"}},
					},
				}
				assert.Equal(t, "getporter/azure-mysql:5.7-interface", dep.Bundle.Interface.Reference, "expected the bundle interface reference to be set")
				assert.Equal(t, wantDoc, dep.Bundle.Interface.Document, "expected the bundle interface document to be set")
			*/
		})
	}
}

func TestAction_Validate_RequireMixinDeclaration(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Sabotage!
	m.Mixins = []MixinDeclaration{}

	err = m.Install.Validate(m)
	assert.EqualError(t, err, "failed to validate 1st step: mixin (exec) was not declared")
}

func TestAction_Validate_RequireMixinData(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Sabotage!
	m.Install[0].Data = nil

	err = m.Install.Validate(m)
	assert.EqualError(t, err, "failed to validate 1st step: no mixin specified")
}

func TestAction_Validate_RequireSingleMixinData(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Sabotage!
	m.Install[0].Data["rando-mixin"] = ""

	err = m.Install.Validate(m)
	assert.EqualError(t, err, "failed to validate 1st step: malformed step, possibly incorrect indentation")
}

func TestAction_Validate_RequireSingleMixinData_Actions(t *testing.T) {
	testcases := []struct {
		name    string
		getStep func(*Manifest) *Steps
	}{
		{"install", func(m *Manifest) *Steps { return &m.Install }},
		{"uninstall", func(m *Manifest) *Steps { return &m.Uninstall }},
		{"upgrade", func(m *Manifest) *Steps { return &m.Upgrade }},
		{"custom", func(m *Manifest) *Steps { status := m.CustomActions["status"]; return &status }},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			c := config.NewTestConfig(t)

			c.TestContext.AddTestFile("testdata/simple.porter.yaml", config.Name)

			m, err := LoadManifestFrom(ctx, c.Config, config.Name)
			require.NoError(t, err, "could not load manifest")
			step := tc.getStep(m)

			if len(*step) == 0 {
				*step = make(Steps, 1)
				(*step)[0] = &Step{
					Data: make(map[string]interface{}),
				}
				(*step)[0].Data["exec"] = ""
			}

			// Sabotage!
			(*step)[0].Data["rando-mixin"] = ""

			err = m.Validate(ctx, c.Config)
			assert.ErrorContains(t, err, "malformed step, possibly incorrect indentation")
		})
	}
}

func TestManifest_Empty_Steps(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/empty-steps.yaml", config.Name)

	_, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	assert.EqualError(t, err, "3 errors occurred:\n\t* validation of action \"install\" failed: failed to validate 2nd step: found an empty step\n\t* validation of action \"uninstall\" failed: failed to validate 2nd step: found an empty step\n\t* validation of action \"status\" failed: failed to validate 1st step: found an empty step\n\n")
}

func TestManifest_Validate_Name(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/porter-no-name.yaml", config.Name)

	_, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	assert.EqualError(t, err, "bundle name must be set")
}

func TestManifest_Validate_Description(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/porter-with-bad-description.yaml", config.Name)

	_, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	assert.ErrorContains(t, err, "validation of action \"install\" failed: failed to validate 1st step: invalid description type (string) for mixin step (exec)")
}

func TestManifest_Validate_InvalidType(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/porter-with-bad-type.yaml", config.Name)

	assert.NotPanics(t, func() {
		_, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
		assert.ErrorContains(t, err, "validation of action \"install\" failed: failed to validate 1st step: invalid mixin type (string) for mixin step (exec)")
	})
}

func TestManifest_Validate_SchemaVersion(t *testing.T) {
	invalidVersionErr := schema.ErrInvalidSchemaVersion.Error()

	t.Run("schemaVersion matches", func(t *testing.T) {
		ctx := context.Background()
		cfg := config.NewTestConfig(t)
		cfg.TestContext.UseFilesystem()
		cfg.Data.SchemaCheck = string(schema.CheckStrategyExact)

		m, err := ReadManifest(cfg.Context, "testdata/porter.yaml")
		require.NoError(t, err)

		err = m.Validate(ctx, cfg.Config)
		require.NoError(t, err)
		assert.NotContains(t, cfg.TestContext.GetError(), invalidVersionErr)
	})

	t.Run("schemaVersion requires experimental feature", func(t *testing.T) {
		ctx := context.Background()
		cfg := config.NewTestConfig(t)
		cfg.TestContext.AddTestFile("testdata/porter.yaml", "porter.yaml")
		cfg.Data.SchemaCheck = string(schema.CheckStrategyExact)

		// Use a schema version that requires dependencies v2 enabled
		cfg.TestContext.EditYaml("porter.yaml", func(yq *yaml.Editor) error {
			return yq.SetValue("schemaVersion", "1.1.0")
		})
		m, err := ReadManifest(cfg.Context, "porter.yaml")
		require.NoError(t, err)

		err = m.Validate(ctx, cfg.Config)
		require.ErrorContains(t, err, "invalid schema version")

		cfg.SetExperimentalFlags(experimental.FlagDependenciesV2)
		err = m.Validate(ctx, cfg.Config)
		require.NoError(t, err)
		assert.NotContains(t, cfg.TestContext.GetError(), invalidVersionErr)
	})

	t.Run("schemaVersion missing, not required", func(t *testing.T) {
		cfg := config.NewTestConfig(t)
		cfg.TestContext.UseFilesystem()
		ctx, span := cfg.StartRootSpan(context.Background(), t.Name())
		defer span.EndSpan()
		cfg.Data.SchemaCheck = string(schema.CheckStrategyNone)

		m, err := ReadManifest(cfg.Context, "testdata/porter.yaml")
		require.NoError(t, err)

		m.SchemaVersion = ""

		err = m.Validate(ctx, cfg.Config)
		require.NoError(t, err)

		// Check that a warning is printed
		// We aren't returning an error because we want to give it a chance to work first. Later we may turn this into a hard error after people have had time to migrate.
		assert.Contains(t, cfg.TestContext.GetError(), invalidVersionErr)
	})
}

func TestManifest_ValidateMetadata(t *testing.T) {
	// Make sure that we allow a range of versions
	invalidVersionErr := schema.ErrInvalidSchemaVersion.Error()
	testcases := []struct {
		schemaVersion string
		wantErr       string
	}{
		{wantErr: invalidVersionErr},
		{schemaVersion: "1.0.0-alpha.1"},
		{schemaVersion: "1.0.0-alpha.2", wantErr: invalidVersionErr},
		{schemaVersion: "1.0.0"},
	}

	for _, tc := range testcases {
		t.Run(tc.schemaVersion, func(t *testing.T) {
			cfg := config.NewTestConfig(t)
			cfg.Data.SchemaCheck = string(schema.CheckStrategyExact)

			m := Manifest{
				SchemaVersion: tc.schemaVersion,
				Name:          "mybuns",
				Registry:      "localhost:5000",
			}
			err := m.validateMetadata(context.Background(), cfg.Config)

			if tc.wantErr == "" {
				require.NoError(t, err)
				assert.NotContains(t, cfg.TestContext.GetError(), invalidVersionErr)
			} else {
				require.ErrorContains(t, err, invalidVersionErr)
			}
		})
	}
}

func TestManifest_ValidateSchemaType(t *testing.T) {
	testcases := []struct {
		schemaType string
		wantErr    string
	}{
		{schemaType: "", wantErr: ""},
		{schemaType: SchemaTypeBundle, wantErr: ""},
		{schemaType: strings.ToLower(SchemaTypeBundle), wantErr: ""},
		{schemaType: strings.ToUpper(SchemaTypeBundle), wantErr: ""},
		{schemaType: "CredentialSet", wantErr: "invalid schemaType CredentialSet, expected Bundle"},
	}

	for _, tc := range testcases {
		t.Run(tc.schemaType, func(t *testing.T) {
			cfg := config.NewTestConfig(t)
			cfg.Data.SchemaCheck = string(schema.CheckStrategyExact)

			m := Manifest{
				SchemaType:    tc.schemaType,
				SchemaVersion: DefaultSchemaVersion.String(),
				Name:          "mybuns",
				Registry:      "localhost:5000",
			}
			err := m.validateMetadata(context.Background(), cfg.Config)

			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
			}
		})
	}
}

func TestManifest_Validate_Dockerfile(t *testing.T) {
	c := config.NewTestConfig(t)
	c.Data.SchemaCheck = string(schema.CheckStrategyNone)

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	m.Dockerfile = "Dockerfile"

	err = m.Validate(context.Background(), c.Config)

	assert.EqualError(t, err, "Dockerfile template cannot be named 'Dockerfile' because that is the filename generated during porter build")
}

func TestManifest_Validate_WrongSchema(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/porter-with-badschema.yaml", config.Name)
	_, err := LoadManifestFrom(context.Background(), c.Config, config.Name)

	require.Error(t, err)
	assert.Regexp(t,
		"unsupported property set or a custom action is defined incorrectly: error unmarshaling custom action baddata",
		err,
	)
}

func TestReadManifest_URL(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	url := "https://raw.githubusercontent.com/getporter/porter/v0.27.1/pkg/manifest/testdata/simple.porter.yaml"
	m, err := ReadManifest(cxt.Context, url)

	require.NoError(t, err)
	assert.Equal(t, "hello", m.Name)
}

func TestReadManifest_Validate_InvalidURL(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	_, err := ReadManifest(cxt.Context, "http://fake-example-porter")

	assert.Error(t, err)
	assert.Regexp(t, "could not reach url http://fake-example-porter", err)
}

func TestReadManifest_File(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)

	require.NoError(t, err)
	assert.Equal(t, "hello", m.Name)
}

func TestSetDefaults(t *testing.T) {
	t.Run("no registry or reference provided", func(t *testing.T) {
		cfg := config.NewTestConfig(t)
		m := Manifest{
			SchemaVersion: DefaultSchemaVersion.String(),
			Name:          "mybun",
			Version:       "1.2.3-beta.1",
		}
		err := m.validateMetadata(context.Background(), cfg.Config)
		require.EqualError(t, err, "a registry or reference value must be provided")
	})

	t.Run("bundle docker tag set on reference", func(t *testing.T) {
		cfg := config.NewTestConfig(t)
		m := Manifest{
			SchemaVersion: DefaultSchemaVersion.String(),
			Name:          "mybun",
			Version:       "1.2.3-beta.1",
			Reference:     "getporter/mybun:v1.2.3",
		}
		err := m.validateMetadata(context.Background(), cfg.Config)
		require.NoError(t, err)

		err = m.SetDefaults()
		require.NoError(t, err)
		assert.Equal(t, "getporter/mybun:v1.2.3", m.Reference)
		assert.Equal(t, "getporter/mybun:porter-e7a4fac8f425d76ed9a5baa3a188824b", m.Image)
	})

	t.Run("bundle docker tag not set on reference", func(t *testing.T) {
		cfg := config.NewTestConfig(t)
		m := Manifest{
			SchemaVersion: DefaultSchemaVersion.String(),
			Name:          "mybun",
			Version:       "1.2.3-beta.1+15",
			Reference:     "getporter/mybun",
		}
		err := m.validateMetadata(context.Background(), cfg.Config)
		require.NoError(t, err)

		err = m.SetDefaults()
		require.NoError(t, err)
		assert.Equal(t, "getporter/mybun:v1.2.3-beta.1_15", m.Reference)
		assert.Equal(t, "getporter/mybun:porter-bcd1325906d287fb3b93500c8bfd2947", m.Image)
	})

	t.Run("bundle reference includes registry with port", func(t *testing.T) {
		cfg := config.NewTestConfig(t)
		m := Manifest{
			SchemaVersion: DefaultSchemaVersion.String(),
			Name:          "mybun",
			Version:       "0.1.0",
			Reference:     "localhost:5000/missing-invocation-image",
		}
		err := m.validateMetadata(context.Background(), cfg.Config)
		require.NoError(t, err)

		err = m.SetDefaults()
		require.NoError(t, err)
		assert.Equal(t, "localhost:5000/missing-invocation-image:v0.1.0", m.Reference)
		assert.Equal(t, "localhost:5000/missing-invocation-image:porter-fea49a80fb6822ee71f71e2ce4a48a37", m.Image)
	})

	t.Run("registry provided, no reference", func(t *testing.T) {
		cfg := config.NewTestConfig(t)
		m := Manifest{
			SchemaVersion: DefaultSchemaVersion.String(),
			Name:          "mybun",
			Version:       "1.2.3-beta.1",
			Registry:      "getporter",
		}
		err := m.validateMetadata(context.Background(), cfg.Config)
		require.NoError(t, err)

		err = m.SetDefaults()
		require.NoError(t, err)
		assert.Equal(t, "getporter/mybun:v1.2.3-beta.1", m.Reference)
		assert.Equal(t, "getporter/mybun:porter-b4b9ce8671aacb5a093574b04f9f87e1", m.Image)
	})

	t.Run("registry provided with org, no reference", func(t *testing.T) {
		cfg := config.NewTestConfig(t)
		m := Manifest{
			SchemaVersion: DefaultSchemaVersion.String(),
			Name:          "mybun",
			Version:       "1.2.3-beta.1",
			Registry:      "getporter/myorg",
		}
		err := m.validateMetadata(context.Background(), cfg.Config)
		require.NoError(t, err)

		err = m.SetDefaults()
		require.NoError(t, err)
		assert.Equal(t, "getporter/myorg/mybun:v1.2.3-beta.1", m.Reference)
		assert.Equal(t, "getporter/myorg/mybun:porter-f4f017f099257ee41d0c05d5e3180f88", m.Image)
	})

	t.Run("registry and reference provided", func(t *testing.T) {
		cfg := config.NewTestConfig(t)
		ctx, span := cfg.StartRootSpan(context.Background(), t.Name()) // Start a span so we can capture trace/logs emitted with a WARNING
		defer span.EndSpan()
		m := Manifest{
			SchemaVersion: DefaultSchemaVersion.String(),
			Name:          "mybun",
			Version:       "1.2.3-beta.1",
			Registry:      "myregistry/myorg",
			Reference:     "getporter/org/mybun:v1.2.3",
		}
		err := m.validateMetadata(ctx, cfg.Config)
		require.NoError(t, err)
		require.Contains(t,
			cfg.TestContext.GetError(),
			"WARNING: both registry and reference were provided; using the reference value of getporter/org/mybun:v1.2.3 for the bundle reference\n",
		)

		err = m.SetDefaults()
		require.NoError(t, err)
		assert.Equal(t, "getporter/org/mybun:v1.2.3", m.Reference)
		assert.Equal(t, "getporter/org/mybun:porter-93d4bfba61358eca91debf6dd4ddc61f", m.Image)
	})
}

func TestReadManifest_Validate_MissingFile(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	_, err := ReadManifest(cxt.Context, "fake-porter.yaml")

	assert.EqualError(t, err, "the specified porter configuration file fake-porter.yaml does not exist")
}

func TestMixinDeclaration_UnmarshalYAML(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.AddTestFile("testdata/mixin-with-config.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)

	require.NoError(t, err)
	assert.Len(t, m.Mixins, 3, "expected 3 mixins")
	assert.Equal(t, "exec", m.Mixins[0].Name)
	assert.Equal(t, "az", m.Mixins[1].Name)
	assert.Equal(t, "terraform", m.Mixins[2].Name)
	assert.Equal(t, map[string]interface{}{"extensions": []interface{}{"iot"}}, m.Mixins[1].Config)
}

func TestMixinDeclaration_UnmarshalYAML_Invalid(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.AddTestFile("testdata/mixin-with-bad-config.yaml", config.Name)
	_, err := ReadManifest(cxt.Context, config.Name)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mixin declaration contained more than one mixin")
}

func TestCredentialsDefinition_UnmarshalYAML(t *testing.T) {
	assertAllCredentialsRequired := func(t *testing.T, creds CredentialDefinitions) {
		for _, cred := range creds {
			assert.EqualValuesf(t, true, cred.Required, "Credential: %s should be required", cred.Name)
		}
	}
	t.Run("all credentials in the generated manifest file are required", func(t *testing.T) {
		cxt := portercontext.NewTestContext(t)
		cxt.AddTestFile("testdata/with-credentials.yaml", config.Name)
		m, err := ReadManifest(cxt.Context, config.Name)
		require.NoError(t, err)
		assertAllCredentialsRequired(t, m.Credentials)

		require.Len(t, m.Credentials, 5)
		assert.Contains(t, m.Credentials, "kubeconfig", "expected a kubeconfig credential definition")
		assert.Equal(t, []string{"status", "uninstall"}, m.Credentials["kubeconfig"].ApplyTo, "credential kubeconfig has incorrect applyTo")

	})
}

func TestMixinDeclaration_MarshalYAML(t *testing.T) {
	m := struct {
		Mixins []MixinDeclaration
	}{
		[]MixinDeclaration{
			{Name: "exec"},
			{Name: "az", Config: map[string]interface{}{"extensions": []interface{}{"iot"}}},
			{Name: "terraform"},
		},
	}

	gotYaml, err := yaml.Marshal(m)
	require.NoError(t, err, "could not marshal data")

	wantYaml, err := os.ReadFile("testdata/mixin-with-config.yaml")
	require.NoError(t, err, "could not read testdata")

	assert.Equal(t, string(wantYaml), string(gotYaml))
}

func TestValidateParameterDefinition_missingPath(t *testing.T) {
	pd := ParameterDefinition{
		Name: "myparam",
		Schema: definition.Schema{
			Type: "file",
		},
	}

	pd.Destination = Location{}

	err := pd.Validate()
	assert.EqualError(t, err, `1 error occurred:
	* no destination path supplied for parameter myparam

`)

	pd.Destination.Path = "/path/to/file"

	err = pd.Validate()
	assert.NoError(t, err)
}

func TestValidateParameterDefinition_invalidSchema(t *testing.T) {
	pd := ParameterDefinition{
		Name: "myparam",
		Schema: definition.Schema{
			Type: "invalid",
		},
	}

	err := pd.Validate()
	assert.Contains(t, err.Error(), `encountered an error while validating definition for parameter "myparam"`)
	assert.Contains(t, err.Error(), `schema not valid: error unmarshaling type from json: "invalid" is not a valid type`)
}

func TestValidateParameterDefinition_defaultFailsValidation(t *testing.T) {
	pd := ParameterDefinition{
		Name: "myparam",
		Schema: definition.Schema{
			Type:    "string",
			Default: 1,
		},
	}

	err := pd.Validate()
	assert.EqualError(t, err, `1 error occurred:
	* encountered an error validating the default value 1 for parameter "myparam": type should be string, got integer

`)
}

func TestValidateOutputDefinition_missingPath(t *testing.T) {
	od := OutputDefinition{
		Name: "myoutput",
		Schema: definition.Schema{
			Type: "file",
		},
	}

	err := od.Validate()
	assert.EqualError(t, err, `1 error occurred:
	* no path supplied for output myoutput

`)

	od.Path = "/path/to/file"

	err = od.Validate()
	assert.NoError(t, err)
}

func TestValidateOutputDefinition_invalidSchema(t *testing.T) {
	od := OutputDefinition{
		Name: "myoutput",
		Schema: definition.Schema{
			Type: "invalid",
		},
	}

	err := od.Validate()
	assert.Contains(t, err.Error(), `encountered an error while validating definition for output "myoutput"`)
	assert.Contains(t, err.Error(), `schema not valid: error unmarshaling type from json: "invalid" is not a valid type`)
}

func TestValidateOutputDefinition_defaultFailsValidation(t *testing.T) {
	od := OutputDefinition{
		Name: "myoutput",
		Schema: definition.Schema{
			Type:    "string",
			Default: 1,
		},
	}

	err := od.Validate()
	assert.EqualError(t, err, `1 error occurred:
	* encountered an error validating the default value 1 for output "myoutput": type should be string, got integer

`)
}

func TestValidateImageMap(t *testing.T) {
	t.Run("with valid image digest, valid repository format and valid tag", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter/myserver",
			Digest:     "sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f",
			Tag:        "latest",
		}

		err := mi.Validate()
		assert.NoError(t, err)
		ref, err := mi.ToOCIReference()
		require.NoError(t, err)
		require.Equal(t, "getporter/myserver@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f", ref.String(), "failed to convert image map to its OCI reference")
	})
	t.Run("with valid repository format and valid tag", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter/myserver",
			Tag:        "v0.1.0",
		}

		err := mi.Validate()
		assert.NoError(t, err)
		ref, err := mi.ToOCIReference()
		require.NoError(t, err)
		require.Equal(t, "getporter/myserver:v0.1.0", ref.String(), "failed to convert image map to its OCI reference")
	})
	t.Run("with both valid image digest and valid repository format", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter/myserver",
			Digest:     "sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f",
		}

		err := mi.Validate()
		require.NoError(t, err)
		ref, err := mi.ToOCIReference()
		require.NoError(t, err)
		require.Equal(t, "getporter/myserver@sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f", ref.String(), "failed to convert image map to its OCI reference")
	})

	t.Run("with no image digest supplied and valid repository format", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter/myserver",
		}

		err := mi.Validate()
		assert.NoError(t, err)
		ref, err := mi.ToOCIReference()
		require.NoError(t, err)
		require.Equal(t, "getporter/myserver", ref.String(), "failed to convert image map to its OCI reference")
	})

	t.Run("with valid image digest but invalid repository format", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter//myserver//",
			Digest:     "sha256:8f1133d81f1b078c865cdb11d17d1ff15f55c449d3eecca50190eed0f5e5e26f",
		}

		err := mi.Validate()
		assert.Error(t, err)
		_, err = mi.ToOCIReference()
		assert.ErrorContains(t, err, "failed to parse named reference")
	})

	t.Run("with invalid image digest format", func(t *testing.T) {
		mi := MappedImage{
			Repository: "getporter/myserver",
			Digest:     "abc123",
		}

		err := mi.Validate()
		assert.Error(t, err)
		_, err = mi.ToOCIReference()
		assert.ErrorContains(t, err, "failed to create a new reference with digest for repository")
	})
}

func TestLoadManifestWithCustomData(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/porter-with-custom-metadata.yaml", config.Name)

	m, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	require.NotNil(t, m, "manifest was nil")
	val, ok := m.Custom["foo"].(map[string]interface{})
	require.True(t, ok, "Cannot cast foo value to map[string]interface{}")

	val1, ok := val["test1"].(bool)
	require.True(t, ok, "Cannot cast test1 value to bool")
	require.True(t, val1, "test1 value is unexpected")

	val2, ok := val["test2"].(int)
	require.True(t, ok, "Cannot cast test2 value to int")
	require.Equal(t, 1, val2, "test2 value is unexpected")

	val3, ok := val["test3"].(string)
	require.True(t, ok, "Cannot cast test3 value to string")
	require.Equal(t, "value", val3, "test3 value is unexpected")

	val4, ok := val["test4"].([]interface{})
	require.True(t, ok, "Cannot cast test4 value to interface{} array")
	val5, ok := val4[0].(string)
	require.True(t, ok, "Cannot cast test4[0] value to string")
	require.Equal(t, "one", val5, "test4[0] value is unexpected")
	val6, ok := val4[1].(string)
	require.True(t, ok, "Cannot cast test4[1] value to string")
	require.Equal(t, "two", val6, "test4[1] value is unexpected")
	val7, ok := val4[2].(string)
	require.True(t, ok, "Cannot cast test4[2] value to string")
	require.Equal(t, "three", val7, "test4[2] value is unexpected")

	val8, ok := val["test5"].(map[string]interface{})
	require.True(t, ok, "Cannot cast test5 value to interface{} array")
	val9, ok := val8["1"].(string)
	require.True(t, ok, "Cannot cast test5[0] value to string")
	require.Equal(t, "one", val9, "test54[0] value is unexpected")
	val10, ok := val8["two"].(string)
	require.True(t, ok, "Cannot cast test5[1] value to string")
	require.Equal(t, "two", val10, "test5[1] value is unexpected")
}

func TestLoadManifestWithRequiredExtensions(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/porter.yaml", config.Name)

	m, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	expected := []RequiredExtension{
		RequiredExtension{
			Name: "requiredExtension1",
		},
		RequiredExtension{
			Name: "requiredExtension2",
			Config: map[string]interface{}{
				"config": true,
			},
		},
	}

	assert.NotNil(t, m)
	assert.Equal(t, expected, m.Required)
}

func TestReadManifest_WithTemplateVariables(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.AddTestFile("testdata/porter-with-templating.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)
	require.NoError(t, err, "ReadManifest failed")
	wantVars := []string{"bundle.dependencies.mysql.outputs.mysql-password", "bundle.outputs.msg", "bundle.outputs.name"}
	assert.Equal(t, wantVars, m.TemplateVariables)
}

func TestManifest_GetTemplatedOutputs(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.AddTestFile("testdata/porter-with-templating.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)
	require.NoError(t, err, "ReadManifest failed")

	outputs := m.GetTemplatedOutputs()

	require.Len(t, outputs, 1)
	assert.Equal(t, "msg", outputs["msg"].Name)
}

func TestManifest_GetTemplatedDependencyOutputs(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.AddTestFile("testdata/porter-with-templating.yaml", config.Name)
	m, err := ReadManifest(cxt.Context, config.Name)
	require.NoError(t, err, "ReadManifest failed")

	outputs := m.GetTemplatedDependencyOutputs()

	require.Len(t, outputs, 1)
	ref := outputs["mysql.mysql-password"]
	assert.Equal(t, "mysql", ref.Dependency)
	assert.Equal(t, "mysql-password", ref.Output)
}

func TestParamToEnvVar(t *testing.T) {
	testcases := []struct {
		name      string
		paramName string
		envName   string
	}{
		{"no special characters", "myparam", "MYPARAM"},
		{"dash", "my-param", "MY_PARAM"},
		{"period", "my.param", "MY_PARAM"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := ParamToEnvVar(tc.paramName)
			assert.Equal(t, tc.envName, got)
		})
	}
}

func TestParameterDefinition_UpdateApplyTo(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	testcases := []struct {
		name         string
		defaultValue string
		applyTo      []string
		source       ParameterSource
		wantApplyTo  []string
	}{
		{"no source", "", nil, ParameterSource{}, nil},
		{"has default", "myparam", nil, ParameterSource{Output: "myoutput"}, nil},
		{"has applyTo", "", []string{"status"}, ParameterSource{Output: "myoutput"}, []string{"status"}},
		{"no default, no applyTo", "", nil, ParameterSource{Output: "myoutput"}, []string{"status", "uninstall"}},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			pd := ParameterDefinition{
				Name: "myparam",
				Schema: definition.Schema{
					Type: "file",
				},
				Source:  tc.source,
				ApplyTo: tc.applyTo,
			}

			if tc.defaultValue != "" {
				pd.Schema.Default = tc.defaultValue
			}

			pd.UpdateApplyTo(m)
			require.Equal(t, tc.wantApplyTo, pd.ApplyTo)
		})
	}
}

func TestManifest_getTemplatePrefix(t *testing.T) {
	testcases := []struct {
		schemaVersion string
		wantPrefix    string
	}{
		{"", ""},
		{"1.0.0-alpha.1", ""},
		{"1.0.0-alpha.2", TemplateDelimiterPrefix},
		{"1.0.0", TemplateDelimiterPrefix},
		{"1.1.0", TemplateDelimiterPrefix},
		{"3.0.0", TemplateDelimiterPrefix},
	}
	for _, tc := range testcases {
		t.Run(tc.schemaVersion, func(t *testing.T) {
			m := Manifest{SchemaVersion: tc.schemaVersion}
			prefix := m.GetTemplatePrefix()
			require.Equal(t, tc.wantPrefix, prefix)
		})
	}
}

func TestManifest_DetermineDependenciesExtensionUsed(t *testing.T) {
	t.Run("no dependencies used", func(t *testing.T) {
		m := Manifest{}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Empty(t, depsExt)
	})

	t.Run("v1 features only", func(t *testing.T) {
		m := Manifest{
			Dependencies: Dependencies{Requires: []*Dependency{
				{
					Name:       "mysql",
					Bundle:     BundleCriteria{Reference: "mysql:5.7", Version: "5.7 - 6"},
					Parameters: map[string]string{"loglevel": "4"},
				},
			}},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV1ExtensionKey, depsExt)
	})

	t.Run("v2 declared but no deps defined", func(t *testing.T) {
		m := Manifest{
			Required: []RequiredExtension{
				{Name: cnab.DependenciesV2ExtensionShortHand},
			},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV2ExtensionKey, depsExt)
	})

	t.Run("v2 shorthand declared", func(t *testing.T) {
		m := Manifest{
			Required: []RequiredExtension{
				{Name: cnab.DependenciesV2ExtensionShortHand},
			},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV2ExtensionKey, depsExt)
	})

	t.Run("v2 full key declared", func(t *testing.T) {
		m := Manifest{
			Required: []RequiredExtension{
				{Name: cnab.DependenciesV2ExtensionKey},
			},
			Dependencies: Dependencies{Requires: []*Dependency{
				{Name: "mysql", Bundle: BundleCriteria{Reference: "mysql:5.7", Version: "5.7 - 6"}},
			}},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV2ExtensionKey, depsExt)
	})

	t.Run("provides interface used", func(t *testing.T) {
		m := Manifest{

			Dependencies: Dependencies{
				Provides: &DependencyProvider{Interface: InterfaceDeclaration{ID: "myinterface"}},
			},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV2ExtensionKey, depsExt)
	})

	t.Run("provides interface empty", func(t *testing.T) {
		// Even if they aren't using it, declaring that the bundle provides an (empty) interface is enough that
		// we should use the v2 dependency
		m := Manifest{
			Dependencies: Dependencies{
				Provides: &DependencyProvider{Interface: InterfaceDeclaration{ID: ""}},
			},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV2ExtensionKey, depsExt)
	})

	t.Run("bundle interface criteria used", func(t *testing.T) {
		m := Manifest{
			Dependencies: Dependencies{Requires: []*Dependency{
				{
					Name: "mysql",
					Bundle: BundleCriteria{
						Reference: "mysql:5.7",
						Version:   "5.7 - 6",
						Interface: &BundleInterface{}}},
			}},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV2ExtensionKey, depsExt)
	})

	t.Run("sharing criteria used", func(t *testing.T) {
		m := Manifest{
			Dependencies: Dependencies{Requires: []*Dependency{
				{
					Name:    "mysql",
					Bundle:  BundleCriteria{Reference: "mysql:5.7", Version: "5.7 - 6"},
					Sharing: SharingCriteria{Mode: true, Group: SharingGroup{Name: "myapp"}},
				},
			}},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV2ExtensionKey, depsExt)
	})

	t.Run("credential wiring used", func(t *testing.T) {
		m := Manifest{
			Dependencies: Dependencies{Requires: []*Dependency{
				{
					Name:        "mysql",
					Bundle:      BundleCriteria{Reference: "mysql:5.7", Version: "5.7 - 6"},
					Credentials: map[string]string{"kubeconfig": "${bundle.credentials.kubeconfig}"},
				},
			}},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV2ExtensionKey, depsExt)
	})

	t.Run("output wiring used", func(t *testing.T) {
		m := Manifest{
			Dependencies: Dependencies{Requires: []*Dependency{
				{
					Name:    "mysql",
					Bundle:  BundleCriteria{Reference: "mysql:5.7", Version: "5.7 - 6"},
					Outputs: map[string]string{"endpoint": "https://${outputs.host}:${outputs.port}/myapp"},
				},
			}},
		}
		depsExt := m.DetermineDependenciesExtensionUsed()
		assert.Equal(t, cnab.DependenciesV2ExtensionKey, depsExt)
	})
}
