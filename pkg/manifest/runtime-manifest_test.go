package manifest

import (
	"fmt"
	"os"
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/deislabs/cnab-go/bundle/definition"
	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/context"
	"github.com/docker/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestLoadManifest(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	require.NotNil(t, m, "manifest was nil")
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
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/porter.yaml", config.Name)
	cxt.AddTestDirectory("testdata/bundles", "bundles")

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	assert.NotNil(t, m)
	assert.Equal(t, []MixinDeclaration{{Name: "exec"}}, m.Mixins)
	assert.Len(t, m.Install, 1)

	installStep := m.Install[0]
	description, _ := installStep.GetDescription()
	assert.NotNil(t, description)

	mixin := installStep.GetMixinName()
	assert.Equal(t, "exec", mixin)
}

func TestAction_Validate_RequireMixinDeclaration(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Sabotage!
	m.Mixins = []MixinDeclaration{}

	err = m.Install.Validate(m)
	assert.EqualError(t, err, "mixin (exec) was not declared")
}

func TestAction_Validate_RequireMixinData(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Sabotage!
	m.Install[0].Data = nil

	err = m.Install.Validate(m)
	assert.EqualError(t, err, "no mixin specified")
}

func TestAction_Validate_RequireSingleMixinData(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	// Sabotage!
	m.Install[0].Data["rando-mixin"] = ""

	err = m.Install.Validate(m)
	assert.EqualError(t, err, "more than one mixin specified")
}

func TestManifest_Validate_Dockerfile(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	m.Dockerfile = "Dockerfile"

	err = m.Validate()

	assert.EqualError(t, err, "Dockerfile template cannot be named 'Dockerfile' because that is the filename generated during porter build")
}

func TestResolveMapParam(t *testing.T) {
	os.Setenv("PERSON", "Ralpha")
	defer os.Unsetenv("PERSON")

	cxt := context.NewTestContext(t)
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "person",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
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

	err = rm.Prepare()
	assert.NoError(t, err)
}

func TestResolvePathParam(t *testing.T) {
	cxt := context.NewTestContext(t)
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
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
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
	cxt := context.NewTestContext(t)
	m := &Manifest{
		Parameters: []ParameterDefinition{},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

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

func TestPrepare_fileParam(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/file-param", "/path/to/file")

	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "file-param",
				Destination: Location{
					Path: "/path/to/file",
				},
				Schema: definition.Schema{
					Type:    "file",
					Default: "/path/to/file",
				},
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Parameters": map[string]interface{}{
				"file-param": "{{bundle.parameters.file-param}}",
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
	val, ok := pms["file-param"].(string)
	assert.True(t, ok)
	assert.Equal(t, "/path/to/file", val)

	err = rm.Prepare()
	assert.NoError(t, err)

	bytes, err := cxt.FileSystem.ReadFile("/path/to/file")
	assert.NoError(t, err)
	assert.Equal(t, "Hello World!", string(bytes), "expected file contents to equal the decoded parameter value")
}

func TestResolveArrayUnknown(t *testing.T) {
	cxt := context.NewTestContext(t)
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "name",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

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

	cxt := context.NewTestContext(t)
	m := &Manifest{
		Parameters: []ParameterDefinition{
			{
				Name: "person",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

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

	cxt := context.NewTestContext(t)
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
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

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

	cxt := context.NewTestContext(t)
	m := &Manifest{
		Credentials: []CredentialDefinition{
			{
				Name:     "password",
				Location: Location{EnvironmentVariable: "PASSWORD"},
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

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
	cxt := context.NewTestContext(t)
	m := &Manifest{
		Dependencies: map[string]Dependency{
			"dep": {
				Tag: "deislabs/porter-hello",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
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

	cxt.FileSystem.WriteFile("/cnab/app/dependencies/dep/outputs/dep_output", []byte("dep_output_value"), 0644)

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
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/param-test-in-block.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

	installStep := rm.Install[0]

	os.Setenv("COMMAND", "echo hello world")
	err = rm.ResolveStep(installStep)
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
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/slice-test.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := RuntimeManifest{Manifest: m}

	installStep := rm.Install[0]

	os.Setenv("COMMAND", "echo hello world")
	err = rm.ResolveStep(installStep)
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

	cxt := context.NewTestContext(t)
	m := &Manifest{
		Mixins: []MixinDeclaration{{Name: "helm"}},
		Install: Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
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

	cxt := context.NewTestContext(t)
	m := &Manifest{
		Mixins: []MixinDeclaration{{Name: "helm"}},
		Install: Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

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

	cxt := context.NewTestContext(t)
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
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

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

	cxt := context.NewTestContext(t)
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
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

	os.Setenv("DATABASE", "wordpress")
	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template values: Missing variable \"nope\"", err.Error())
}

func TestManifest_ResolveBundleName(t *testing.T) {
	cxt := context.NewTestContext(t)
	m := &Manifest{
		Name: "mybundle",
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)

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
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/outputs/bundle-outputs.yaml", config.Name)

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

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	require.Equal(t, wantOutputs, m.Outputs)
}

func TestReadManifest_Validate_BundleOutput_Error(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/outputs/bundle-outputs-error.yaml", config.Name)

	_, err := LoadManifestFrom(cxt.Context, config.Name)
	require.Error(t, err)
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
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/simple.porter.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := RuntimeManifest{Manifest: m}

	depStep := rm.Install[0]
	err = rm.ApplyStepOutputs(depStep, map[string]string{"foo": "bar"})
	require.NoError(t, err)

	assert.Contains(t, rm.outputs, "foo")
	assert.Equal(t, "bar", rm.outputs["foo"])
}

func makeBoolPtr(value bool) *bool {
	return &value
}

func TestManifest_ResolveImageMap(t *testing.T) {
	cxt := context.NewTestContext(t)
	cxt.AddTestFile("testdata/porter-images.yaml", config.Name)

	m, err := LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := RuntimeManifest{Manifest: m}
	expectedImage, ok := m.ImageMap["something"]
	require.True(t, ok, "couldn't get expected image")
	expectedRef := fmt.Sprintf("%s@%s", expectedImage.Repository, expectedImage.Digest)
	step := rm.Install[0]
	err = rm.ResolveStep(step)
	assert.NoError(t, err, "Should have successfully resolved step")
	s := step.Data["searcher"].(map[interface{}]interface{})
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
	assert.True(t, ok, "should have found digest")
	val = fmt.Sprintf("%v", digest)
	assert.Equal(t, expectedImage.Digest, val)

	tag, ok := s["tag"]
	assert.True(t, ok, "should have found tag")
	val = fmt.Sprintf("%v", tag)
	assert.Equal(t, expectedImage.Tag, val)
}

func TestManifest_ResolveImageMapMissingKey(t *testing.T) {

	cxt := context.NewTestContext(t)
	m := &Manifest{
		Name: "mybundle",
		ImageMap: map[string]MappedImage{
			"something": MappedImage{
				Repository: "blah/blah",
				Digest:     "sha1234:cafebab",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step exercising bundle image interpolation",
			"Arguments": []string{
				"{{ bundle.images.something.Fake }}",
			},
		},
	}
	err := rm.ResolveStep(s)
	assert.Error(t, err)
}

func TestManifest_ResolveImageMapMissingImage(t *testing.T) {

	cxt := context.NewTestContext(t)
	m := &Manifest{
		Name: "mybundle",
		ImageMap: map[string]MappedImage{
			"notsomething": MappedImage{
				Repository: "blah/blah",
				Digest:     "sha1234:cafebab",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
	s := &Step{
		Data: map[string]interface{}{
			"description": "a test step exercising bundle image interpolation",
			"Arguments": []string{
				"{{ bundle.images.something.Fake }}",
			},
		},
	}
	err := rm.ResolveStep(s)
	assert.Error(t, err)
}

func TestResolveImage(t *testing.T) {
	tests := []struct {
		name      string
		reference string
		want      MappedImage
	}{
		{
			name:      "canonical reference",
			reference: "deislabs/porter-hello@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want: MappedImage{
				Repository: "deislabs/porter-hello",
				Digest:     "sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			},
		},
		{
			name:      "tagged reference",
			reference: "deislabs/porter-hello:v0.1.10",
			want: MappedImage{
				Repository: "deislabs/porter-hello",
				Tag:        "v0.1.10",
			},
		},
		{
			name:      "named reference",
			reference: "deislabs/porter-hello",
			want: MappedImage{
				Repository: "deislabs/porter-hello",
				Tag:        "latest",
			},
		},
		{
			name:      "the one with a hostname",
			reference: "deislabs.io/deislabs/porter-hello",
			want: MappedImage{
				Repository: "deislabs.io/deislabs/porter-hello",
				Tag:        "latest",
			},
		},
		{
			name:      "the one with a hostname and port",
			reference: "deislabs.io:9090/deislabs/porter-hello:foo",
			want: MappedImage{
				Repository: "deislabs.io:9090/deislabs/porter-hello",
				Tag:        "foo",
			},
		},
		{

			name:      "tagged and digested",
			reference: "deislabs/porter-hello:latest@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want: MappedImage{
				Repository: "deislabs/porter-hello",
				Tag:        "latest",
				Digest:     "sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			},
		},
	}
	for _, test := range tests {
		got := &MappedImage{}
		err := resolveImage(got, test.reference)
		assert.NoError(t, err)
		assert.Equal(t, test.want.Repository, got.Repository)
		assert.Equal(t, test.want.Tag, got.Tag)
		assert.Equal(t, test.want.Digest, got.Digest)
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
			reference: "deislabs/porter-hello@8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want:      "unable to parse docker image %s: invalid reference format",
		},
		{
			name:      "bad digest",
			reference: "deislabs/porter-hello@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f",
			want:      "unable to parse docker image %s: invalid checksum digest length",
		},
		{
			name:      "bad digest algo",
			reference: "deislabs/porter-hello@sha356:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want:      "unable to parse docker image %s: unsupported digest algorithm",
		},
		{
			name:      "malformed tagged ref",
			reference: "deislabs/porter-hello@latest",
			want:      "unable to parse docker image %s: invalid reference format",
		},
		{
			name:      "too many ports tagged ref",
			reference: "deislabs:8080:8080/porter-hello:latest",
			want:      "unable to parse docker image %s: invalid reference format",
		},
	}
	for _, test := range tests {
		got := &MappedImage{}
		err := resolveImage(got, test.reference)
		assert.EqualError(t, err, fmt.Sprintf(test.want, test.reference))
	}
}

func TestResolveImageWithUpdatedBundle(t *testing.T) {
	cxt := context.NewTestContext(t)
	m := &Manifest{
		ImageMap: map[string]MappedImage{
			"machine": MappedImage{
				Repository: "deislabs/ghost",
				Tag:        "latest",
				Digest:     "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "blah/ghost:latest"
	img.Digest = "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"
	bun := &bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	}

	reloMap := relocation.ImageRelocationMap{}

	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "blah/ghost", mi.Repository)
}

func TestResolveImageWithUpdatedMismatchedBundle(t *testing.T) {
	cxt := context.NewTestContext(t)
	m := &Manifest{
		ImageMap: map[string]MappedImage{
			"machine": MappedImage{
				Repository: "deislabs/ghost",
				Tag:        "latest",
				Digest:     "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "blah/ghost:latest"
	img.Digest = "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"
	bun := &bundle.Bundle{
		Images: map[string]bundle.Image{
			"ghost": img,
		},
	}

	reloMap := relocation.ImageRelocationMap{}

	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("unable to find image in porter manifest: %s", "ghost"))

}

func TestResolveImageWithRelo(t *testing.T) {
	cxt := context.NewTestContext(t)
	m := &Manifest{
		ImageMap: map[string]MappedImage{
			"machine": MappedImage{
				Repository: "gabrtv/microservice",
				Tag:        "latest",
				Digest:     "sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687"
	img.Digest = "sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687"
	bun := &bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	}

	reloMap := relocation.ImageRelocationMap{
		"gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687": "my.registry/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687",
	}

	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "my.registry/microservice", mi.Repository)
}

func TestResolveImageBadRelocation(t *testing.T) {
	cxt := context.NewTestContext(t)
	m := &Manifest{
		ImageMap: map[string]MappedImage{
			"machine": MappedImage{
				Repository: "deislabs/ghost",
				Tag:        "latest",
				Digest:     "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041",
			},
		},
	}

	img := bundle.Image{}
	img.Image = "deislabs/ghost:latest"
	img.Digest = "sha256:75c495e5ce9c428d482973d72e3ce9925e1db304a97946c9aa0b540d7537e041"
	bun := &bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	}

	reloMap := relocation.ImageRelocationMap{
		"deislabs/nogood:latest": "cnabio/ghost:latest",
	}

	rm := NewRuntimeManifest(cxt.Context, ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("unable to find relocated image: %s", "deislabs/nogood:latest"))
}
