package runtime

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/experimental"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/tests"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runtimeManifestFromStepYaml(t *testing.T, testConfig *config.TestConfig, stepYaml string) *RuntimeManifest {
	mContent := []byte(stepYaml)
	require.NoError(t, testConfig.FileSystem.WriteFile("/cnab/app/porter.yaml", mContent, pkg.FileModeWritable))
	m, err := manifest.ReadManifest(testConfig.Context, "/cnab/app/porter.yaml", testConfig.Config)
	require.NoError(t, err, "ReadManifest failed")
	cfg := NewConfigFor(testConfig.Config)
	return NewRuntimeManifest(cfg, cnab.ActionInstall, m)
}

func TestResolveMapParam(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("PERSON", "Ralpha")
	testConfig.Setenv("CONTACT", "{ \"name\": \"Breta\" }")

	mContent := `schemaVersion: 1.0.0-alpha.2
parameters:
- name: person
- name: place
  applyTo: [install]
- name: contact
  type: object

install:
- mymixin:
    Parameters:
      Thing: ${ bundle.parameters.person }
      ObjectName: ${ bundle.parameters.contact.name }
      Object: '${ bundle.parameters.contact }'
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Parameters"], map[string]interface{}{}, "Data.mymixin.Parameters has incorrect type")
	pms := mixin["Parameters"].(map[string]interface{})
	require.IsType(t, "string", pms["Thing"], "Data.mymixin.Parameters.Thing has incorrect type")
	val := pms["Thing"].(string)

	assert.Equal(t, "Ralpha", val)
	assert.NotContains(t, "place", pms, "parameters that don't apply to the current action should not be resolved")

	// Asserting `bundle.parameters.contact.name` works.
	require.IsType(t, "string", pms["ObjectName"], "Data.mymixin.Parameters.ObjectName has incorrect type")
	contactName := pms["ObjectName"].(string)
	require.IsType(t, "string", contactName, "Data.mymixin.Parameters.ObjectName.name has incorrect type")
	assert.Equal(t, "Breta", contactName)

	// Asserting `bundle.parameters.contact` evaluates to the JSON string
	// representation of the object.
	require.IsType(t, "string", pms["Object"], "Data.mymixin.Parameters.Object has incorrect type")
	contact := pms["Object"].(string)
	require.IsType(t, "string", contact, "Data.mymixin.Parameters.Object has incorrect type")
	assert.Equal(t, "{\"name\":\"Breta\"}", contact)

	err = rm.Initialize(ctx)
	require.NoError(t, err)
}
func TestStateBagUnpack(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("PERSON", "Ralpha")

	mContent := `schemaVersion: 1.0.0-alpha.2
parameters:
- name: person
- name: place
  applyTo: [install]

install:
- mymixin:
    Parameters:
      Thing: ${ bundle.parameters.person }
state:
- name: foo
  path: foo/state.json
`
	tests := []struct {
		name         string
		stateContent string
		expErr       error
	}{
		{
			name:         "/porter/state.tgz is empty file",
			stateContent: "",
			expErr:       nil,
		},
		{
			name:         "/porter/state.tgz has null string",
			stateContent: "null",
			expErr:       nil,
		},
		{
			name:         "/porter/state.tgz has newline",
			stateContent: "\n",
			expErr:       io.ErrUnexpectedEOF,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
			require.NoError(t, testConfig.FileSystem.WriteFile("/porter/state.tgz", []byte(test.stateContent), pkg.FileModeWritable))
			s := rm.Install[0]

			err := rm.ResolveStep(ctx, 0, s)
			require.NoError(t, err)

			err = rm.Initialize(ctx)
			if test.expErr == nil {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), test.expErr.Error())
			}
			if test.stateContent != "null" {
				err = testConfig.FileSystem.Remove("/porter/state.tgz")
				require.NoError(t, err)
			}
		})
	}
}

func TestResolvePathParam(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)

	mContent := `schemaVersion: 1.0.0-alpha.2
parameters:
- name: person
  path: person.txt

install:
- mymixin:
    Parameters:
      Thing: ${ bundle.parameters.person }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Parameters"], map[string]interface{}{}, "Data.mymixin.Parameters has incorrect type")
	pms := mixin["Parameters"].(map[string]interface{})
	require.IsType(t, "string", pms["Thing"], "Data.mymixin.Parameters.Thing has incorrect type")
	val := pms["Thing"].(string)

	assert.Equal(t, "person.txt", val)
}

func TestMetadataAvailableForTemplating(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/metadata-substitution.yaml", config.Name)
	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "LoadManifestFrom")
	cfg := NewConfigFor(c.Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

	s := rm.Install[0]
	err = rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	pms, ok := s.Data["exec"].(map[string]interface{})
	require.True(t, ok)
	cmd := pms["command"].(string)
	assert.Equal(t, "echo \"name:porter-hello version:0.1.0 description:An example Porter configuration image:jeremyrickard/porter-hello:porter-39a022ca907e26c3d8fffabd4bb8dbbc\"", cmd)
}

func TestDependencyMetadataAvailableForTemplating(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/dep-metadata-substitution.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "LoadManifestFrom")
	cfg := NewConfigFor(c.Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	rm.bundles = map[string]cnab.ExtendedBundle{
		"mysql": cnab.NewBundle(bundle.Bundle{
			Name:        "Azure MySQL",
			Description: "Azure MySQL database as a service",
			Version:     "v1.0.0",
		}),
	}

	s := rm.Install[0]
	err = rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	pms, ok := s.Data["exec"].(map[string]interface{})
	require.True(t, ok)
	cmd := pms["command"].(string)
	assert.Equal(t, "echo \"dep name: Azure MySQL dep version: v1.0.0 dep description: Azure MySQL database as a service\"", cmd)
}

func TestResolveMapParamUnknown(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)

	mContent := `schemaVersion: 1.0.0
install:
- mymixin:
    Parameters:
      Thing: ${bundle.parameters.person}
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.Error(t, err)
	tests.RequireErrorContains(t, err, "missing variable \"person\"")
}

func TestResolveArrayUnknown(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)

	mContent := `schemaVersion: 1.0.0
parameters:
- name: name

install:
- exec:
    Arguments:
      - ${bundle.parameters.person}
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `missing variable "person"`)
}

func TestResolveArray(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("PERSON", "Ralpha")

	mContent := `schemaVersion: 1.0.0
parameters:
- name: person

install:
- mymixin:
    Arguments:
    - ${ bundle.parameters.person }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, "Ralpha", args[0].(string))
}

func TestResolveSensitiveParameter(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("SENSITIVE_PARAM", "deliciou$dubonnet")
	testConfig.Setenv("SENSITIVE_OBJECT", "{ \"secret\": \"this_is_secret\" }")
	testConfig.Setenv("REGULAR_PARAM", "regular param value")

	mContent := `schemaVersion: 1.0.0
parameters:
- name: sensitive_param
  sensitive: true
- name: sensitive_object
  sensitive: true
  type: object
- name: regular_param

install:
- mymixin:
    Arguments:
    - ${ bundle.parameters.sensitive_param }
    - '${ bundle.parameters.sensitive_object }'
    - ${ bundle.parameters.sensitive_object.secret }
    - ${ bundle.parameters.regular_param }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	// Prior to resolving step values, this method should return an empty string array
	assert.Equal(t, rm.GetSensitiveValues(), []string{})

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	require.Len(t, args, 4)
	assert.Equal(t, "deliciou$dubonnet", args[0])
	assert.Equal(t, "{\"secret\":\"this_is_secret\"}", args[1])
	assert.Equal(t, "this_is_secret", args[2])
	assert.Equal(t, "regular param value", args[3])

	// Verify sensitive values include both the whole object and sub-properties
	sensitiveValues := rm.GetSensitiveValues()
	assert.Contains(t, sensitiveValues, "deliciou$dubonnet")
	assert.Contains(t, sensitiveValues, "{\"secret\":\"this_is_secret\"}")
	assert.Contains(t, sensitiveValues, "this_is_secret", "sub-property value should be tracked as sensitive")
}

func TestResolveSensitiveObjectSubProperties(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("SENSITIVE_CREDS", `{"apiKey": "secret-key-123", "password": "secret-pass-456", "nested": {"token": "nested-token-789"}}`)
	testConfig.Setenv("SENSITIVE_ARRAY", `{"keys": ["key1", "key2"], "count": 42}`)

	mContent := `schemaVersion: 1.0.0
parameters:
- name: sensitive_creds
  sensitive: true
  type: object
- name: sensitive_array
  sensitive: true
  type: object

install:
- mymixin:
    Arguments:
    - ${ bundle.parameters.sensitive_creds.apiKey }
    - ${ bundle.parameters.sensitive_creds.password }
    - ${ bundle.parameters.sensitive_creds.nested.token }
    - '${ bundle.parameters.sensitive_creds }'
    - '${ bundle.parameters.sensitive_creds.nested }'
    - ${ bundle.parameters.sensitive_array.count }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	require.Len(t, args, 6)
	assert.Equal(t, "secret-key-123", args[0])
	assert.Equal(t, "secret-pass-456", args[1])
	assert.Equal(t, "nested-token-789", args[2])
	// When accessing a nested object property, it resolves to the map representation
	assert.Equal(t, "map[token:nested-token-789]", args[4])

	// Verify all sub-property values are tracked as sensitive
	sensitiveValues := rm.GetSensitiveValues()
	assert.Contains(t, sensitiveValues, "secret-key-123", "apiKey value should be tracked as sensitive")
	assert.Contains(t, sensitiveValues, "secret-pass-456", "password value should be tracked as sensitive")
	assert.Contains(t, sensitiveValues, "nested-token-789", "nested token value should be tracked as sensitive")
	assert.Contains(t, sensitiveValues, "key1", "array element should be tracked as sensitive")
	assert.Contains(t, sensitiveValues, "key2", "array element should be tracked as sensitive")
	assert.Contains(t, sensitiveValues, "42", "numeric value should be tracked as sensitive")

	// Verify intermediate nested objects are also tracked
	assert.Contains(t, sensitiveValues, `{"token":"nested-token-789"}`, "nested object should be tracked as sensitive")

	// The whole object JSON strings should also be tracked
	assert.Contains(t, sensitiveValues, `{"apiKey":"secret-key-123","nested":{"token":"nested-token-789"},"password":"secret-pass-456"}`)
}

func TestResolveCredential(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("PASSWORD", "deliciou$dubonnet")

	mContent := `schemaVersion: 1.0.0
credentials:
- name: password
  env: PASSWORD

install:
- mymixin:
    Arguments:
    - ${ bundle.credentials.password }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	// Prior to resolving step values, this method should return an empty string array
	assert.Equal(t, rm.GetSensitiveValues(), []string{})

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, "deliciou$dubonnet", args[0])
	// There should now be a sensitive value tracked under the manifest
	assert.Equal(t, []string{"deliciou$dubonnet"}, rm.GetSensitiveValues())
}

func TestResolveStep_DependencyOutput(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("PORTER_MYSQL_PASSWORD_DEP_OUTPUT", "password")
	testConfig.Setenv("PORTER_MYSQL_ROOT_PASSWORD_DEP_OUTPUT", "mysql-password")

	mContent := `schemaVersion: 1.0.0
dependencies:
  requires: 
  - name: mysql
    bundle:
      reference: "getporter/porter-mysql"

install:
- mymixin:
    Arguments:
    - ${ bundle.dependencies.mysql.outputs.password }
    - ${ bundle.dependencies.mysql.outputs.root-password }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	ps := cnab.ParameterSources{}
	ps.SetParameterFromDependencyOutput("porter-mysql-password", "mysql", "password")
	ps.SetParameterFromDependencyOutput("porter-mysql-root-password", "mysql", "root-password")
	rm.bundle = cnab.NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			cnab.ParameterSourcesExtensionKey: ps,
		},
		RequiredExtensions: []string{cnab.ParameterSourcesExtensionKey},
	})

	rm.bundles = map[string]cnab.ExtendedBundle{
		"mysql": cnab.NewBundle(bundle.Bundle{
			Outputs: map[string]bundle.Output{
				"password": {
					Definition: "password",
				},
				"root-password": {
					Definition: "root-password",
				},
			},
			Definitions: map[string]*definition.Schema{
				"password":      {WriteOnly: makeBoolPtr(true)},
				"root-password": {WriteOnly: makeBoolPtr(true)},
			},
		}),
	}

	s := rm.Install[0]
	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, []interface{}{"password", "mysql-password"}, args, "Incorrect template args passed to the mixin step")

	// There should now be a sensitive value tracked under the manifest
	gotSensitiveValues := rm.GetSensitiveValues()
	sort.Strings(gotSensitiveValues)
	assert.Equal(t, []string{"mysql-password", "password"}, gotSensitiveValues, "Incorrect values were marked as sensitive")
}

func TestResolveStep_DependencyMappedOutput(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.SetExperimentalFlags(experimental.FlagDependenciesV2)

	mContent := `schemaVersion: 1.0.0
dependencies:
  requires: 
  - name: mysql
    bundle:
      reference: "getporter/porter-mysql"
    outputs:
      mappedOutput: Mapped

install:
- mymixin:
    Arguments:
    - ${ bundle.dependencies.mysql.outputs.mappedOutput }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	rm.bundles = map[string]cnab.ExtendedBundle{
		"mysql": cnab.NewBundle(bundle.Bundle{}),
	}

	s := rm.Install[0]
	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, []interface{}{"Mapped"}, args, "Incorrect template args passed to the mixin step")
}

func TestResolveStep_DependencyTemplatedMappedOutput(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.SetExperimentalFlags(experimental.FlagDependenciesV2)
	testConfig.Setenv("PORTER_MYSQL_PASSWORD_DEP_OUTPUT", "password")

	mContent := `schemaVersion: 1.0.0
dependencies:
  requires: 
  - name: mysql
    bundle:
      reference: "getporter/porter-mysql"
    outputs:
      mappedOutput: ${ bundle.dependencies.mysql.outputs.password }

install:
- mymixin:
    Arguments:
    - ${ bundle.dependencies.mysql.outputs.mappedOutput }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	ps := cnab.ParameterSources{}
	ps.SetParameterFromDependencyOutput("porter-mysql-password", "mysql", "password")
	rm.bundle = cnab.NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			cnab.ParameterSourcesExtensionKey: ps,
		},
		RequiredExtensions: []string{cnab.ParameterSourcesExtensionKey},
	})

	rm.bundles = map[string]cnab.ExtendedBundle{
		"mysql": cnab.NewBundle(bundle.Bundle{
			Outputs: map[string]bundle.Output{
				"password": {
					Definition: "password",
				},
			},
			Definitions: map[string]*definition.Schema{
				"password": {WriteOnly: makeBoolPtr(true)},
			},
		}),
	}

	s := rm.Install[0]
	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, []interface{}{"password"}, args, "Incorrect template args passed to the mixin step")
}

func TestResolveStep_DependencyTemplatedMappedOutput_OutputVariable(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.SetExperimentalFlags(experimental.FlagDependenciesV2)
	testConfig.Setenv("PORTER_MYSQL_PASSWORD_DEP_OUTPUT", "password")

	mContent := `schemaVersion: 1.0.0
dependencies:
  requires: 
  - name: mysql
    bundle:
      reference: "getporter/porter-mysql"
    outputs:
      mappedOutput: combined-${ outputs.password }

install:
- mymixin:
    Arguments:
    - ${ bundle.dependencies.mysql.outputs.mappedOutput }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	ps := cnab.ParameterSources{}
	ps.SetParameterFromDependencyOutput("porter-mysql-password", "mysql", "password")
	rm.bundle = cnab.NewBundle(bundle.Bundle{
		Custom: map[string]interface{}{
			cnab.ParameterSourcesExtensionKey: ps,
		},
		RequiredExtensions: []string{cnab.ParameterSourcesExtensionKey},
	})

	rm.bundles = map[string]cnab.ExtendedBundle{
		"mysql": cnab.NewBundle(bundle.Bundle{
			Outputs: map[string]bundle.Output{
				"password": {
					Definition: "password",
				},
			},
			Definitions: map[string]*definition.Schema{
				"password": {WriteOnly: makeBoolPtr(true)},
			},
		}),
	}

	s := rm.Install[0]
	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has incorrect type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, mixin["Arguments"], []interface{}{}, "Data.mymixin.Arguments has incorrect type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, []interface{}{"combined-password"}, args, "Incorrect template args passed to the mixin step")
}

func TestResolveInMainDict(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/param-test-in-block.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	cfg := NewConfigFor(c.Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

	installStep := rm.Install[0]

	rm.config.Setenv("COMMAND", "echo hello world")
	err = rm.ResolveStep(ctx, 0, installStep)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, installStep.Data["exec"], "Data.exec has the wrong type")
	exec := installStep.Data["exec"].(map[string]interface{})
	command := exec["command"]
	require.IsType(t, "string", command, "Data.exec.command has the wrong type")
	cmdVal := command.(string)

	assert.Equal(t, "echo hello world", cmdVal)
}

func TestResolveSliceWithAMap(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/slice-test.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	cfg := NewConfigFor(c.Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

	installStep := rm.Install[0]

	rm.config.Setenv("COMMAND", "echo hello world")
	err = rm.ResolveStep(ctx, 0, installStep)
	require.NoError(t, err)

	require.NotNil(t, installStep.Data)
	exec := installStep.Data["exec"].(map[string]interface{})
	require.NotNil(t, exec)
	flags := exec["flags"].(map[string]interface{})
	require.Len(t, flags, 1)
	assert.Equal(t, "echo hello world", flags["c"].(string))
}

func TestResolveMissingStepOutputs(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)

	mContent := `schemaVersion: 1.0.0
install:
- mymixin:
    Arguments:
    - jdbc://${bundle.outputs.database_url}:${bundle.outputs.database_port}
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	tests.RequireErrorContains(t, err, `missing variable "database_url"`)
}

func TestResolveSensitiveOutputs(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	mContent := `schemaVersion: 1.0.0
outputs:
- name: username
- name: password
  sensitive: true

install:
- mymixin:
    Arguments:
    - ${ bundle.outputs.username }
    - ${ bundle.outputs.password }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	rm.outputs = map[string]string{
		"username": "sally",
		"password": "top$ecret!",
	}
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, s.Data["mymixin"], map[string]interface{}{}, "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, []interface{}{}, mixin["Arguments"], "Data.mymixin.Arguments has the wrong type")
	args := mixin["Arguments"].([]interface{})

	require.Len(t, args, 2)
	require.Equal(t, "sally", args[0])
	require.Equal(t, "top$ecret!", args[1])

	// There should be only one sensitive value being tracked
	require.Equal(t, []string{"top$ecret!"}, rm.GetSensitiveValues())
}

func TestManifest_ResolveBundleName(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	mContent := `schemaVersion: 1.0.0
name: mybuns

install:
- mymixin:
    Arguments:
    - ${ bundle.name }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, s.Data["mymixin"], map[string]interface{}{}, "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, []interface{}{}, mixin["Arguments"], "Data.mymixin.Arguments has the wrong type")
	args := mixin["Arguments"].([]interface{})

	assert.Equal(t, "mybuns", args[0].(string))
}

func TestReadManifest_Validate_BundleOutput(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/outputs/bundle-outputs.yaml", config.Name)

	wantOutputs := manifest.OutputDefinitions{
		"mysql-root-password": {
			Name: "mysql-root-password",
			Schema: definition.Schema{
				Description: "The root MySQL password",
				Type:        "string",
			},
		},
		"mysql-password": {
			Name: "mysql-password",
			Schema: definition.Schema{
				Type: "string",
			},
			ApplyTo: []string{
				"install",
				"upgrade",
			},
		},
	}

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	require.Equal(t, wantOutputs, m.Outputs)
}

func TestReadManifest_Validate_BundleOutput_Error(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/outputs/bundle-outputs-error.yaml", config.Name)

	_, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.Error(t, err)
}

func TestDependencyV1_Validate(t *testing.T) {
	testcases := []struct {
		name       string
		dep        manifest.Dependency
		wantOutput string
		wantError  string
	}{
		{
			name:       "version in reference",
			dep:        manifest.Dependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql:5.7"}},
			wantOutput: "",
			wantError:  "",
		}, {
			name:       "version ranges",
			dep:        manifest.Dependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql", Version: "5.7.x-6"}},
			wantOutput: "",
			wantError:  "",
		}, {
			name:       "missing reference",
			dep:        manifest.Dependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: ""}},
			wantOutput: "",
			wantError:  `reference is required for dependency "mysql"`,
		}, {
			name:       "version not specified",
			dep:        manifest.Dependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql", Version: ""}},
			wantOutput: "",
			wantError:  `reference for dependency "mysql" can specify only a repository, without a digest or tag, when a version constraint is specified`,
		}, { // When a range is specified, but also a default version, we use the default version when we can't find a matching version from the range
			name:       "default version and range specified",
			dep:        manifest.Dependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql:5.7", Version: "5.7.x-6"}},
			wantOutput: "",
			wantError:  "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			pCtx := portercontext.NewTestContext(t)

			err := tc.dep.Validate(pCtx.Context)

			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				tests.RequireErrorContains(t, err, tc.wantError)
			}

			gotOutput := pCtx.GetOutput()
			if gotOutput != "" {
				require.Equal(t, tc.wantOutput, gotOutput)
			}
		})
	}
}

func TestManifest_ApplyStepOutputs(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFileFromRoot("pkg/manifest/testdata/porter-with-templating.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	cfg := NewConfigFor(c.Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

	err = rm.ApplyStepOutputs(map[string]string{"name": "world"})
	require.NoError(t, err)

	assert.Contains(t, rm.outputs, "name")
	assert.Equal(t, "world", rm.outputs["name"])
}

func makeBoolPtr(value bool) *bool {
	return &value
}

func TestManifest_ResolveImageMap(t *testing.T) {
	ctx := context.Background()
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/porter-images.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	cfg := NewConfigFor(c.Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	expectedImage, ok := m.ImageMap["something"]
	require.True(t, ok, "couldn't get expected image")
	expectedRef := fmt.Sprintf("%s@%s", expectedImage.Repository, expectedImage.Digest)
	step := rm.Install[0]
	err = rm.ResolveStep(ctx, 0, step)
	assert.NoError(t, err, "Should have successfully resolved step")
	s := step.Data["searcher"].(map[string]interface{})
	assert.NotNil(t, s)
	img, ok := s["image"]
	assert.True(t, ok, "should have found image")
	val := fmt.Sprintf("%v", img)
	assert.Equal(t, expectedRef, val)

	repo, ok := s["repo"]
	assert.True(t, ok, "should have found repo")
	val = fmt.Sprintf("%v", repo)
	assert.Equal(t, expectedImage.Repository, val)

	digest, ok := s["digest"]
	assert.True(t, ok, "should have found content digest")
	val = fmt.Sprintf("%v", digest)
	assert.Equal(t, expectedImage.Digest, val)

	tag, ok := s["tag"]
	assert.True(t, ok, "should have found tag")
	val = fmt.Sprintf("%v", tag)
	assert.Equal(t, expectedImage.Tag, val)
}

func TestManifest_ResolveImageMapMissingKey(t *testing.T) {
	// Try to access an images entry that doesn't exist
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	mContent := `schemaVersion: 1.0.0-alpha.2
images:
  something:
    repository: "blah/blah"
    digest: "sha1234:cafebab"

install:
- mymixin:
    Arguments:
      - ${ bundle.images.notsomething.digest }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	tests.RequireErrorContains(t, err, `missing variable "notsomething"`)
}

func TestResolveImage(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		want      manifest.MappedImage
	}{
		{
			name:      "canonical reference",
			reference: "getporter/porter-hello@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want: manifest.MappedImage{
				Repository: "getporter/porter-hello",
				Digest:     "sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			},
		},
		{
			name:      "tagged reference",
			reference: "ghcr.io/getporter/examples/porter-hello:v0.2.0",
			want: manifest.MappedImage{
				Repository: "ghcr.io/getporter/examples/porter-hello",
				Tag:        "v0.2.0",
			},
		},
		{
			name:      "named reference",
			reference: "getporter/porter-hello",
			want: manifest.MappedImage{
				Repository: "getporter/porter-hello",
				Tag:        "latest",
			},
		},
		{
			name:      "the one with a hostname",
			reference: "deislabs.io/getporter/porter-hello",
			want: manifest.MappedImage{
				Repository: "deislabs.io/getporter/porter-hello",
				Tag:        "latest",
			},
		},
		{
			name:      "the one with a hostname and port",
			reference: "deislabs.io:9090/getporter/porter-hello:foo",
			want: manifest.MappedImage{
				Repository: "deislabs.io:9090/getporter/porter-hello",
				Tag:        "foo",
			},
		},
		{

			name:      "tagged and digested",
			reference: "ghcr.io/getporter/examples/porter-hello:v0.2.0@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want: manifest.MappedImage{
				Repository: "ghcr.io/getporter/examples/porter-hello",
				Tag:        "v0.2.0",
				Digest:     "sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := &manifest.MappedImage{}
			err := resolveImage(got, test.reference)
			require.NoError(t, err)
			assert.Equal(t, test.want.Repository, got.Repository)
			assert.Equal(t, test.want.Tag, got.Tag)
			assert.Equal(t, test.want.Digest, got.Digest)
		})
	}
}

func TestResolveImageErrors(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		want      string
	}{
		{
			name:      "no algo digest",
			reference: "getporter/porter-hello@8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want:      "invalid reference format",
		},
		{
			name:      "bad digest",
			reference: "getporter/porter-hello@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f",
			want:      "invalid checksum digest length",
		},
		{
			name:      "bad digest algo",
			reference: "getporter/porter-hello@sha356:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want:      "unsupported digest algorithm",
		},
		{
			name:      "malformed tagged ref",
			reference: "getporter/porter-hello@latest",
			want:      "invalid reference format",
		},
		{
			name:      "too many ports tagged ref",
			reference: "deislabs:8080:8080/porter-hello:latest",
			want:      "invalid reference format",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := &manifest.MappedImage{}
			err := resolveImage(got, test.reference)
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.want)
		})
	}
}

func TestResolveImageWithUpdatedBundle(t *testing.T) {
	m := &manifest.Manifest{
		ImageMap: map[string]manifest.MappedImage{
			"machine": manifest.MappedImage{
				Repository: "deislabs/ghost",
				Tag:        "latest",
				Digest:     "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "blah/ghost:latest"
	img.Digest = "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"
	bun := cnab.NewBundle(bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	})

	reloMap := relocation.ImageRelocationMap{}

	cfg := NewConfigFor(config.NewTestConfig(t).Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	require.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "blah/ghost", mi.Repository)
}

func TestResolveImageWithUpdatedMismatchedBundle(t *testing.T) {
	m := &manifest.Manifest{
		ImageMap: map[string]manifest.MappedImage{
			"machine": manifest.MappedImage{
				Repository: "deislabs/ghost",
				Tag:        "latest",
				Digest:     "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "blah/ghost:latest"
	img.Digest = "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"
	bun := cnab.NewBundle(bundle.Bundle{
		Images: map[string]bundle.Image{
			"ghost": img,
		},
	})

	reloMap := relocation.ImageRelocationMap{}

	cfg := NewConfigFor(config.NewTestConfig(t).Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("unable to find image in porter manifest: %s", "ghost"))

}

func TestResolveImageWithRelo(t *testing.T) {
	m := &manifest.Manifest{
		ImageMap: map[string]manifest.MappedImage{
			"machine": manifest.MappedImage{
				Repository: "gabrtv/microservice",
				Tag:        "latest",
				Digest:     "sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687"
	img.Digest = "sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687"
	bun := cnab.NewBundle(bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	})

	reloMap := relocation.ImageRelocationMap{
		"gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687": "my.registry/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687",
	}

	cfg := NewConfigFor(config.NewTestConfig(t).Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	require.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "my.registry/microservice", mi.Repository)
}

func TestResolveImageRelocationNoMatch(t *testing.T) {
	m := &manifest.Manifest{
		ImageMap: map[string]manifest.MappedImage{
			"machine": manifest.MappedImage{
				Repository: "deislabs/ghost",
				Tag:        "latest",
				Digest:     "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "deislabs/ghost:latest"
	img.Digest = "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"
	bun := cnab.NewBundle(bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	})

	reloMap := relocation.ImageRelocationMap{
		"deislabs/nogood:latest": "cnabio/ghost:latest",
	}

	cfg := NewConfigFor(config.NewTestConfig(t).Config)
	rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	require.NoError(t, err)
	assert.Equal(t, "deislabs/ghost", rm.ImageMap["machine"].Repository)
}

func TestResolveStepEncoding(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)

	wantValue := `{"test":"value"}`
	testConfig.Setenv("TEST", wantValue)

	mContent := `schemaVersion: 1.0.0
parameters:
- name: test
  env: TEST

install:
- mymixin:
    Flags:
      c: '${bundle.parameters.test}'
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, s.Data["mymixin"], map[string]interface{}{}, "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})
	require.IsType(t, map[string]interface{}{}, mixin["Flags"], "Data.mymixin.Flags has the wrong type")
	flags := mixin["Flags"].(map[string]interface{})

	assert.Equal(t, flags["c"], wantValue)
}

func TestResolveInstallation(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv(config.EnvPorterInstallationNamespace, "mynamespace")
	testConfig.Setenv(config.EnvPorterInstallationName, "mybun")
	testConfig.Setenv(config.EnvPorterInstallationID, "myid")

	mContent := `schemaVersion: 1.0.0
install:
- mymixin:
    ns: ${ installation.namespace }
    release: ${ installation.name }
    id: ${ installation.id }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})

	assert.Equal(t, "mynamespace", mixin["ns"], "installation.namespace was not rendered")
	assert.Equal(t, "mybun", mixin["release"], "installation.name was not rendered")
	assert.Equal(t, "myid", mixin["id"], "installation.id was not rendered")
}

func TestResolveCustomMetadata(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)

	mContent := `schemaVersion: 1.0.0
custom:
  foo: foobar
  myApp:
    featureFlags:
      featureA: true

install:
- mymixin:
    release: ${ bundle.custom.foo }
    featureA: ${ bundle.custom.myApp.featureFlags.featureA }
    notabool: "${ bundle.custom.myApp.featureFlags.featureA }"
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})

	err = rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err, "ResolveStep failed")

	assert.Equal(t, "foobar", mixin["release"], "custom metadata was not rendered")
	assert.Equal(t, true, mixin["featureA"], "nested custom metadata was not rendered, an unquoted boolean should render as a bool")
	assert.Equal(t, "true", mixin["notabool"], "a quoted boolean should render as a string")
}

func TestResolveEnvironmentVariable(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)
	testConfig.Setenv("foo", "foo-value")
	testConfig.Setenv("BAR", "bar-value")

	mContent := `schemaVersion: 1.0.0
install:
- mymixin:
    someInput: ${ env.foo }
    moreInput: ${ env.BAR }
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)
	s := rm.Install[0]

	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err)

	require.IsType(t, map[string]interface{}{}, s.Data["mymixin"], "Data.mymixin has the wrong type")
	mixin := s.Data["mymixin"].(map[string]interface{})

	assert.Equal(t, "foo-value", mixin["someInput"], "expected lower-case foo env var was resolved")
	assert.Equal(t, "bar-value", mixin["moreInput"], "expected upper-case BAR env var was resolved")
}

func TestStateBagUnpackWithParentDirectoryCreation(t *testing.T) {
	ctx := context.Background()
	testConfig := config.NewTestConfig(t)

	mContent := `schemaVersion: 1.0.0-alpha.2
state:
- name: some_state
  path: /path/to/some_state

install:
- mymixin:
    Parameters:
      Thing: test
`
	rm := runtimeManifestFromStepYaml(t, testConfig, mContent)

	// Create a valid tar.gz file with state content
	stateContent := "test state content"
	stateArchive := createTestStateArchive(t, "some_state", stateContent)

	// Write the state archive to the expected location
	require.NoError(t, testConfig.FileSystem.WriteFile("/porter/state.tgz", stateArchive, pkg.FileModeWritable))

	// Resolve the step first (following the pattern from existing tests)
	s := rm.Install[0]
	err := rm.ResolveStep(ctx, 0, s)
	require.NoError(t, err, "ResolveStep should succeed")

	// Test that Initialize succeeds and creates the parent directory
	err = rm.Initialize(ctx)
	if err != nil {
		t.Logf("Initialize failed with error: %v", err)
		t.Logf("StateBag length: %d", len(rm.StateBag))
		for i, state := range rm.StateBag {
			t.Logf("StateBag[%d]: name=%s, path=%s", i, state.Name, state.Path)
		}
	}
	require.NoError(t, err, "Initialize should succeed and create parent directories for state files")

	// Verify that the parent directory was created
	parentDir := "/path/to"
	exists, err := testConfig.FileSystem.DirExists(parentDir)
	require.NoError(t, err)
	assert.True(t, exists, "Parent directory should be created")

	// Verify that the state file was created with correct content
	stateFilePath := "/path/to/some_state"
	exists, err = testConfig.FileSystem.DirExists(stateFilePath)
	require.NoError(t, err)
	assert.False(t, exists, "State file should not be a directory")

	content, err := testConfig.FileSystem.ReadFile(stateFilePath)
	require.NoError(t, err)
	assert.Equal(t, stateContent, string(content), "State file should contain the expected content")

	// Clean up
	require.NoError(t, testConfig.FileSystem.Remove("/porter/state.tgz"))
}

// createTestStateArchive creates a tar.gz archive with the given state file
func createTestStateArchive(t *testing.T, stateName, content string) []byte {
	// Create a buffer to write our archive to
	var buf bytes.Buffer

	// Create a gzip writer
	gw := gzip.NewWriter(&buf)
	defer gw.Close()

	// Create a tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Create tar header
	header := &tar.Header{
		Name: "porter-state/" + stateName,
		Mode: 0644,
		Size: int64(len(content)),
	}

	// Write the header
	err := tw.WriteHeader(header)
	require.NoError(t, err)

	// Write the content
	_, err = tw.Write([]byte(content))
	require.NoError(t, err)

	// Close the writers to flush the data
	tw.Close()
	gw.Close()

	return buf.Bytes()
}

func TestResolveInvocationImage(t *testing.T) {
	testcases := []struct {
		name                string
		bundleInvocationImg bundle.BaseImage
		relocationMap       relocation.ImageRelocationMap
		expectedImg         string
		wantErr             string
	}{
		{name: "success with no relocation map",
			bundleInvocationImg: bundle.BaseImage{Image: "blah/ghost:latest", ImageType: "docker", Digest: "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"},
			expectedImg:         "blah/ghost:latest@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
		},
		{name: "success with relocation map",
			bundleInvocationImg: bundle.BaseImage{Image: "blah/ghost:latest", ImageType: "docker", Digest: "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"},
			relocationMap:       relocation.ImageRelocationMap{"blah/ghost:latest@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041": "relocated-ghost@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"},
			expectedImg:         "relocated-ghost@sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
		},
		{name: "success with no update",
			expectedImg: "test/image:latest",
		},
		{name: "failure with invalid digest",
			bundleInvocationImg: bundle.BaseImage{Image: "blah/ghost:latest", ImageType: "docker", Digest: "123"},
			wantErr:             "unable to get bundle image reference with digest",
		},
	}

	cfg := NewConfigFor(config.NewTestConfig(t).Config)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			bun := cnab.NewBundle(bundle.Bundle{
				InvocationImages: []bundle.InvocationImage{
					{BaseImage: tc.bundleInvocationImg},
				},
			})
			m := &manifest.Manifest{
				Image: "test/image:latest",
			}
			rm := NewRuntimeManifest(cfg, cnab.ActionInstall, m)

			err := rm.ResolveInvocationImage(bun, tc.relocationMap)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedImg, m.Image)
		})
	}

}
