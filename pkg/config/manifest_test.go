package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestReadManifest_URL(t *testing.T) {
	c := NewTestConfig(t)
	url := "https://raw.githubusercontent.com/deislabs/porter/master/pkg/config/testdata/simple.porter.yaml"
	m, err := c.ReadManifest(url)

	require.NoError(t, err)
	assert.Equal(t, "hello", m.Name)
	assert.Equal(t, url, m.path)
}

func TestReadManifest_Validate_InvalidURL(t *testing.T) {
	c := NewTestConfig(t)
	_, err := c.ReadManifest("http://fake-example-porter")

	assert.Error(t, err)
	assert.Regexp(t, "could not reach url http://fake-example-porter", err)
}

func TestReadManifest_File(t *testing.T) {
	c := NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)
	m, err := c.ReadManifest(Name)

	require.NoError(t, err)
	assert.Equal(t, "hello", m.Name)
	assert.Equal(t, Name, m.path)
}

func TestReadManifest_Validate_MissingFile(t *testing.T) {
	c := NewTestConfig(t)
	_, err := c.ReadManifest("fake-porter.yaml")

	assert.EqualError(t, err, "the specified porter configuration file fake-porter.yaml does not exist")
}

func TestLoadManifest(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	require.NoError(t, c.LoadManifest(), "could not load manifest")

	require.NotNil(t, c.Manifest, "manifest was nil")
	assert.Equal(t, []string{"exec"}, c.Manifest.Mixins, "expected manifest to declare the exec mixin")
	require.Len(t, c.Manifest.Install, 1, "expected 1 install step")

	installStep := c.Manifest.Install[0]
	description, _ := installStep.GetDescription()
	assert.NotNil(t, description, "expected the install description to be populated")

	mixin := installStep.GetMixinName()
	assert.Equal(t, "exec", mixin, "incorrect install step mixin used")

	require.Len(t, c.Manifest.CustomActions, 1, "expected manifest to declare 1 custom action")
	require.Contains(t, c.Manifest.CustomActions, "status", "expected manifest to declare a status action")

	statusStep := c.Manifest.CustomActions["status"][0]
	description, _ = statusStep.GetDescription()
	assert.Equal(t, "Get World Status", description, "unexpected status step description")

	mixin = statusStep.GetMixinName()
	assert.Equal(t, "exec", mixin, "unexpected status step mixin")
}

func TestLoadManifestWithDependencies(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)
	c.TestContext.AddTestDirectory("testdata/bundles", "bundles")

	require.NoError(t, c.LoadManifest())

	assert.NotNil(t, c.Manifest)
	assert.Equal(t, []string{"exec"}, c.Manifest.Mixins)
	assert.Len(t, c.Manifest.Install, 1)

	installStep := c.Manifest.Install[0]
	description, _ := installStep.GetDescription()
	assert.NotNil(t, description)

	mixin := installStep.GetMixinName()
	assert.Equal(t, "exec", mixin)
}

func TestAction_Validate_RequireMixinDeclaration(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	// Sabotage!
	c.Manifest.Mixins = []string{}

	err = c.Manifest.Install.Validate(c.Manifest)
	assert.EqualError(t, err, "mixin (exec) was not declared")
}

func TestAction_Validate_RequireMixinData(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	// Sabotage!
	c.Manifest.Install[0].Data = nil

	err = c.Manifest.Install.Validate(c.Manifest)
	assert.EqualError(t, err, "no mixin specified")
}

func TestAction_Validate_RequireSingleMixinData(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	// Sabotage!
	c.Manifest.Install[0].Data["rando-mixin"] = ""

	err = c.Manifest.Install.Validate(c.Manifest)
	assert.EqualError(t, err, "more than one mixin specified")
}

func TestManifest_Validate_Dockerfile(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	err := c.LoadManifest()
	require.NoError(t, err)

	c.Manifest.Dockerfile = "Dockerfile"

	err = c.Manifest.Validate()

	assert.EqualError(t, err, "Dockerfile template cannot be named 'Dockerfile' because that is the filename generated during porter build")
}

func TestResolveMapParam(t *testing.T) {
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "person",
			},
		},
	}

	os.Setenv("PERSON", "Ralpha")
	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Parameters": map[string]interface{}{
				"Thing": "{{bundle.parameters.person}}",
			},
		},
	}

	before, _ := yaml.Marshal(s)
	t.Logf("Before:\n %s", before)
	err := m.ResolveStep(s)
	require.NoError(t, err)
	after, _ := yaml.Marshal(s)
	t.Logf("After:\n %s", after)
	assert.NotNil(t, s.Data)
	t.Logf("Length of data:%d", len(s.Data))
	assert.NotEmpty(t, s.Data["Parameters"])
	for k, v := range s.Data {
		t.Logf("Key %s, value: %s, type: %T", k, v, v)
	}
	pms, ok := s.Data["Parameters"].(map[interface{}]interface{})
	assert.True(t, ok)
	val, ok := pms["Thing"].(string)
	assert.True(t, ok)
	assert.Equal(t, "Ralpha", val)
}

func TestResolveMapParamUnknown(t *testing.T) {

	m := &Manifest{
		Parameters: []ParameterDefinition{},
	}

	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Parameters": map[string]interface{}{
				"Thing": "{{bundle.parameters.person}}",
			},
		},
	}

	err := m.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template values: Missing variable \"person\"", err.Error())
}

func TestResolveArrayUnknown(t *testing.T) {
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "name",
			},
		},
	}

	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.parameters.person }}",
			},
		},
	}

	err := m.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template values: Missing variable \"person\"", err.Error())
}

func TestResolveArray(t *testing.T) {
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "person",
			},
		},
	}

	os.Setenv("PERSON", "Ralpha")
	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.parameters.person }}",
			},
		},
	}

	err := m.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Ralpha", args[0].(string))
}

func TestResolveSensitiveParameter(t *testing.T) {
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name:      "sensitive_param",
				Sensitive: true,
			},
			{
				Name: "regular_param",
			},
		},
	}

	os.Setenv("SENSITIVE_PARAM", "deliciou$dubonnet")
	os.Setenv("REGULAR_PARAM", "regular param value")
	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.parameters.sensitive_param }}",
				"{{ bundle.parameters.regular_param }}",
			},
		},
	}

	// Prior to resolving step values, this method should return an empty string array
	assert.Equal(t, m.GetSensitiveValues(), []string{})

	err := m.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(args))
	assert.Equal(t, "deliciou$dubonnet", args[0])
	assert.Equal(t, "regular param value", args[1])

	// There should now be one sensitive value tracked under the manifest
	assert.Equal(t, []string{"deliciou$dubonnet"}, m.GetSensitiveValues())
}

func TestResolveCredential(t *testing.T) {
	m := &Manifest{
		Credentials: []CredentialDefinition{
			{
				Name:     "password",
				Location: Location{EnvironmentVariable: "PASSWORD"},
			},
		},
	}

	os.Setenv("PASSWORD", "deliciou$dubonnet")
	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.credentials.password }}",
			},
		},
	}

	// Prior to resolving step values, this method should return an empty string array
	assert.Equal(t, m.GetSensitiveValues(), []string{})

	err := m.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "deliciou$dubonnet", args[0])

	// There should now be a sensitive value tracked under the manifest
	assert.Equal(t, []string{"deliciou$dubonnet"}, m.GetSensitiveValues())
}

func TestResolveStepOutputs(t *testing.T) {
	t.Skip("Skip while dependencies is being rewritten")

	m := &Manifest{
		outputs: map[string]string{
			"output": "output_value",
		},
		Dependencies: map[string]Dependency{
			"dep": {
				Tag: "deislabs/porter-hello",
			},
		},
	}

	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.outputs.output }}",
				"{{ bundle.dependencies.dep.outputs.dep_output }}",
			},
		},
	}

	// Prior to resolving step values, this method should return an empty string array
	assert.Equal(t, m.GetSensitiveValues(), []string{})

	err := m.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(args))
	assert.Equal(t, "output_value", args[0].(string))
	assert.Equal(t, "dep_output_value", args[1].(string))

	// There should now be a sensitive value tracked under the manifest
	assert.Equal(t, []string{"output_value", "dep_output_value"}, m.GetSensitiveValues())
}

func TestResolveInMainDict(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/param-test-in-block.yaml", Name)

	require.NoError(t, c.LoadManifest())

	installStep := c.Manifest.Install[0]

	os.Setenv("COMMAND", "echo hello world")
	err := c.Manifest.ResolveStep(installStep)
	assert.NoError(t, err)

	assert.NotNil(t, installStep.Data)
	t.Logf("install data %v", installStep.Data)
	exec := installStep.Data["exec"].(map[interface{}]interface{})
	assert.NotNil(t, exec)
	command := exec["command"].(interface{})
	assert.NotNil(t, command)
	cmdVal, ok := command.(string)
	assert.True(t, ok)
	assert.Equal(t, "echo hello world", cmdVal)
}

func TestResolveSliceWithAMap(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/slice-test.yaml", Name)

	require.NoError(t, c.LoadManifest())

	installStep := c.Manifest.Install[0]

	os.Setenv("COMMAND", "echo hello world")
	err := c.Manifest.ResolveStep(installStep)
	assert.NoError(t, err)

	assert.NotNil(t, installStep.Data)
	t.Logf("install data %v", installStep.Data)
	exec := installStep.Data["exec"].(map[interface{}]interface{})
	assert.NotNil(t, exec)
	args := exec["arguments"].([]interface{})
	assert.Len(t, args, 2)
	assert.Equal(t, "echo hello world", args[1].(string))
	assert.NotNil(t, args)
}

func TestResolveMultipleStepOutputs(t *testing.T) {

	databaseURL := "localhost"
	databasePort := "3303"

	s := &Step{
		Data: map[string]interface{}{
			"helm": map[interface{}]interface{}{
				"description": "install wordpress",
				"Arguments": []string{
					"jdbc://{{bundle.outputs.database_url}}:{{bundle.outputs.database_port}}",
				},
			},
		},
	}

	m := &Manifest{
		Mixins: []string{"helm"},
		Install: Steps{
			s,
		},
		outputs: map[string]string{
			"database_url":  databaseURL,
			"database_port": databasePort,
		},
	}

	err := m.ResolveStep(s)
	require.NoError(t, err)
	helm, ok := s.Data["helm"].(map[interface{}]interface{})
	assert.True(t, ok)
	args, ok := helm["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, fmt.Sprintf("jdbc://%s:%s", databaseURL, databasePort), args[0].(string))
}

func TestResolveMissingStepOutputs(t *testing.T) {

	s := &Step{
		Data: map[string]interface{}{
			"helm": map[interface{}]interface{}{
				"description": "install wordpress",
				"Arguments": []string{
					"jdbc://{{bundle.outputs.database_url}}:{{bundle.outputs.database_port}}",
				},
			},
		},
	}

	m := &Manifest{
		Mixins: []string{"helm"},
		Install: Steps{
			s,
		},
	}

	err := m.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template values: Missing variable \"database_url\"", err.Error())
}

func TestResolveDependencyParam(t *testing.T) {
	t.Skip("Skip while dependencies is being rewritten")

	s := &Step{
		Data: map[string]interface{}{
			"helm": map[interface{}]interface{}{
				"description": "install wordpress",
				"Arguments": []string{
					"{{bundle.dependencies.mysql.parameters.database}}",
				},
			},
		},
	}

	m := &Manifest{
		Dependencies: map[string]Dependency{
			"mysql": {
				Tag: "deislabs/porter-mysql",
			},
		},
		Mixins: []string{"helm"},
		Install: Steps{
			s,
		},
	}

	os.Setenv("DATABASE", "wordpress")
	err := m.ResolveStep(s)
	require.NoError(t, err)
	helm, ok := s.Data["helm"].(map[interface{}]interface{})
	assert.True(t, ok)
	args, ok := helm["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "wordpress", args[0].(string))
}

func TestResolveMissingDependencyParam(t *testing.T) {
	t.Skip("Skip while dependencies is being rewritten")

	s := &Step{
		Data: map[string]interface{}{
			"helm": map[interface{}]interface{}{
				"description": "install wordpress",
				"Arguments": []string{
					"{{bundle.dependencies.mysql.parameters.nope}}",
				},
			},
		},
	}

	m := &Manifest{
		Dependencies: map[string]Dependency{
			"mysql": {
				Tag: "deislabs/porter-mysql",
			},
		},
		Mixins: []string{"helm"},
		Install: Steps{
			s,
		},
	}

	os.Setenv("DATABASE", "wordpress")
	err := m.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template values: Missing variable \"nope\"", err.Error())
}

func TestManifest_ApplyBundleOutputs(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	require.NoError(t, c.LoadManifest())

	depStep := c.Manifest.Install[0]
	err := c.Manifest.ApplyStepOutputs(depStep, map[string]string{"foo": "bar"})
	require.NoError(t, err)

	assert.Contains(t, c.Manifest.outputs, "foo")
	assert.Equal(t, "bar", c.Manifest.outputs["foo"])
}

func TestManifest_ResolveBundleName(t *testing.T) {
	m := &Manifest{
		Name: "mybundle",
	}

	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step exercising bundle name interpolation",
			"Arguments": []string{
				"{{ bundle.name }}",
			},
		},
	}

	err := m.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "mybundle", args[0].(string))
}

func TestReadManifest_Validate_BundleOutput(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/outputs/bundle-outputs.yaml", Name)

	wantOutputs := []OutputDefinition{
		{
			Name:        "mysql-root-password",
			Description: "The root MySQL password",
			Schema: Schema{
				Type: "string",
			},
		},
		{
			Name: "mysql-password",
			Schema: Schema{
				Type: "string",
			},
			ApplyTo: []string{
				"install",
				"upgrade",
			},
		},
	}

	require.NoError(t, c.LoadManifest())
	require.NoError(t, c.Manifest.Validate())
	require.Equal(t, wantOutputs, c.Manifest.Outputs)
}

func TestReadManifest_Validate_BundleOutput_Error(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/outputs/bundle-outputs-error.yaml", Name)

	require.Error(t, c.LoadManifest())
}

func TestDependency_Validate(t *testing.T) {
	testcases := []struct {
		name      string
		dep       Dependency
		wantError string
	}{
		{"version in tag", Dependency{Tag: "deislabs/azure-mysql:5.7"}, ""},
		{"version ranges", Dependency{Tag: "deislabs/azure-mysql", Versions: []string{"5.7.x-6"}}, ""},
		{"missing tag", Dependency{Tag: ""}, "dependency tag is required"},
		{"version double specified", Dependency{Tag: "deislabs/azure-mysql:5.7", Versions: []string{"5.7.x-6"}}, "dependency tag can only specify REGISTRY/NAME when version ranges are specified"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dep.Validate()

			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tc.wantError)
			}
		})
	}
}
