package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/porter/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestLoadManifest(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	require.NoError(t, c.LoadManifest(), "could not load manifest")

	require.NotNil(t, c.Manifest, "manifest was nil")
	assert.Equal(t, []MixinDeclaration{{Name: "exec"}}, c.Manifest.Mixins, "expected manifest to declare the exec mixin")
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
	assert.Equal(t, []MixinDeclaration{{Name: "exec"}}, c.Manifest.Mixins)
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
	c.Manifest.Mixins = []MixinDeclaration{}

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
	os.Setenv("PERSON", "Ralpha")
	defer os.Unsetenv("PERSON")

	c := NewTestConfig(t)
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "person",
			},
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)
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
	err := rm.ResolveStep(s)
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

func TestResolvePathParam(t *testing.T) {

	c := NewTestConfig(t)
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "person",
				Destination: Location{
					Path: "person.txt",
				},
			},
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)
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
	err := rm.ResolveStep(s)
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
	assert.Equal(t, "person.txt", val)
}

func TestResolveMapParamUnknown(t *testing.T) {
	c := context.NewTestContext(t)
	m := &Manifest{
		Parameters: []ParameterDefinition{},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)

	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Parameters": map[string]interface{}{
				"Thing": "{{bundle.parameters.person}}",
			},
		},
	}

	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template Parameters:\n  Thing: '{{bundle.parameters.person}}'\ndescription: a test step\n: Missing variable \"person\"", err.Error())
}

func TestResolveArrayUnknown(t *testing.T) {
	c := NewTestConfig(t)
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "name",
			},
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)

	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.parameters.person }}",
			},
		},
	}

	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template Arguments:\n- '{{ bundle.parameters.person }}'\ndescription: a test step\n: Missing variable \"person\"", err.Error())
}

func TestResolveArray(t *testing.T) {
	os.Setenv("PERSON", "Ralpha")
	defer os.Unsetenv("PERSON")

	c := NewTestConfig(t)
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "person",
			},
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)

	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.parameters.person }}",
			},
		},
	}

	err := rm.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Ralpha", args[0].(string))
}

func TestResolveSensitiveParameter(t *testing.T) {
	os.Setenv("SENSITIVE_PARAM", "deliciou$dubonnet")
	defer os.Unsetenv("SENSITIVE_PARAM")
	os.Setenv("REGULAR_PARAM", "regular param value")
	defer os.Unsetenv("REGULAR_PARAM")

	c := NewTestConfig(t)
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
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)

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
	assert.Equal(t, rm.GetSensitiveValues(), []string{})

	err := rm.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(args))
	assert.Equal(t, "deliciou$dubonnet", args[0])
	assert.Equal(t, "regular param value", args[1])

	// There should now be one sensitive value tracked under the manifest
	assert.Equal(t, []string{"deliciou$dubonnet"}, rm.GetSensitiveValues())
}

func TestResolveCredential(t *testing.T) {
	os.Setenv("PASSWORD", "deliciou$dubonnet")
	defer os.Unsetenv("PASSWORD")

	c := NewTestConfig(t)
	m := &Manifest{
		Credentials: []CredentialDefinition{
			{
				Name:     "password",
				Location: Location{EnvironmentVariable: "PASSWORD"},
			},
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)

	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.credentials.password }}",
			},
		},
	}

	// Prior to resolving step values, this method should return an empty string array
	assert.Equal(t, rm.GetSensitiveValues(), []string{})

	err := rm.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "deliciou$dubonnet", args[0])

	// There should now be a sensitive value tracked under the manifest
	assert.Equal(t, []string{"deliciou$dubonnet"}, rm.GetSensitiveValues())
}

func TestResolveStepOutputs(t *testing.T) {
	c := NewTestConfig(t)
	m := &Manifest{
		Dependencies: map[string]Dependency{
			"dep": {
				Tag: "deislabs/porter-hello",
			},
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)
	rm.bundles = map[string]bundle.Bundle{
		"dep": {
			Outputs: map[string]bundle.Output{
				"dep_output": {
					Definition: "dep_output",
				},
			},
			Definitions: map[string]*definition.Schema{
				"dep_output": {WriteOnly: makeBoolPtr(true)},
			},
		},
	}
	rm.outputs = map[string]string{
		"output": "output_value",
	}

	c.FileSystem.WriteFile("/cnab/app/dependencies/dep/outputs/dep_output", []byte(`{"value":"dep_output_value"}`), 0644)

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
	assert.Equal(t, rm.GetSensitiveValues(), []string{})

	err := rm.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(args))
	assert.Equal(t, "output_value", args[0].(string))
	assert.Equal(t, "dep_output_value", args[1].(string))

	// There should now be a sensitive value tracked under the manifest
	assert.Equal(t, []string{"output_value", "dep_output_value"}, rm.GetSensitiveValues())
}

func TestResolveInMainDict(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/param-test-in-block.yaml", Name)

	require.NoError(t, c.LoadManifest())
	rm := NewRuntimeManifest(c.Context, ActionInstall, c.Manifest)

	installStep := rm.Install[0]

	os.Setenv("COMMAND", "echo hello world")
	err := rm.ResolveStep(installStep)
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
	rm := RuntimeManifest{Manifest: c.Manifest}

	installStep := rm.Install[0]

	os.Setenv("COMMAND", "echo hello world")
	err := rm.ResolveStep(installStep)
	assert.NoError(t, err)

	assert.NotNil(t, installStep.Data)
	t.Logf("install data %v", installStep.Data)
	exec := installStep.Data["exec"].(map[interface{}]interface{})
	assert.NotNil(t, exec)
	flags := exec["flags"].(map[interface{}]interface{})
	assert.Len(t, flags, 1)
	assert.Equal(t, "echo hello world", flags["c"].(string))
	assert.NotNil(t, flags)
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

	c := NewTestConfig(t)
	m := &Manifest{
		Mixins: []MixinDeclaration{{Name: "helm"}},
		Install: Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)
	rm.outputs = map[string]string{
		"database_url":  databaseURL,
		"database_port": databasePort,
	}

	err := rm.ResolveStep(s)
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

	c := NewTestConfig(t)
	m := &Manifest{
		Mixins: []MixinDeclaration{{Name: "helm"}},
		Install: Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)

	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template helm:\n  Arguments:\n  - jdbc://{{bundle.outputs.database_url}}:{{bundle.outputs.database_port}}\n  description: install wordpress\n: Missing variable \"database_url\"", err.Error())
}

func TestResolveDependencyParam(t *testing.T) {
	t.Skip("still haven't decided if this is going to be supported")

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

	c := NewTestConfig(t)
	m := &Manifest{
		Dependencies: map[string]Dependency{
			"mysql": {
				Tag: "deislabs/porter-mysql",
			},
		},
		Mixins: []MixinDeclaration{{Name: "helm"}},
		Install: Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)

	os.Setenv("DATABASE", "wordpress")
	err := rm.ResolveStep(s)
	require.NoError(t, err)
	helm, ok := s.Data["helm"].(map[interface{}]interface{})
	assert.True(t, ok)
	args, ok := helm["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, "wordpress", args[0].(string))
}

func TestResolveMissingDependencyParam(t *testing.T) {
	t.Skip("still haven't decided if this is going to be supported")

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

	c := NewTestConfig(t)
	m := &Manifest{
		Dependencies: map[string]Dependency{
			"mysql": {
				Tag: "deislabs/porter-mysql",
			},
		},
		Mixins: []MixinDeclaration{{Name: "helm"}},
		Install: Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)

	os.Setenv("DATABASE", "wordpress")
	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template values: Missing variable \"nope\"", err.Error())
}

func TestManifest_ResolveBundleName(t *testing.T) {
	c := context.NewTestContext(t)
	m := &Manifest{
		Name: "mybundle",
	}
	rm := NewRuntimeManifest(c.Context, ActionInstall, m)

	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step exercising bundle name interpolation",
			"Arguments": []string{
				"{{ bundle.name }}",
			},
		},
	}

	err := rm.ResolveStep(s)
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
			Name: "mysql-root-password",
			Schema: definition.Schema{
				Description: "The root MySQL password",
				Type:        "string",
			},
		},
		{
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

func TestManifest_ApplyStepOutputs(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	require.NoError(t, c.LoadManifest())
	rm := RuntimeManifest{Manifest: c.Manifest}

	depStep := rm.Install[0]
	err := rm.ApplyStepOutputs(depStep, map[string]string{"foo": "bar"})
	require.NoError(t, err)

	assert.Contains(t, rm.outputs, "foo")
	assert.Equal(t, "bar", rm.outputs["foo"])
}

func makeBoolPtr(value bool) *bool {
	return &value
}
