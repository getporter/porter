package runtime

import (
	"fmt"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/context"
	"get.porter.sh/porter/pkg/manifest"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestResolveMapParam(t *testing.T) {
	os.Setenv("PERSON", "Ralpha")
	defer os.Unsetenv("PERSON")

	cxt := context.NewTestContext(t)
	m := &manifest.Manifest{
		Parameters: []manifest.ParameterDefinition{
			{
				Name: "person",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	s := &manifest.Step{
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
	m := &manifest.Manifest{
		Parameters: []manifest.ParameterDefinition{
			{
				Name: "person",
				Destination: manifest.Location{
					Path: "person.txt",
				},
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	s := &manifest.Step{
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

func TestMetadataAvailableForTemplating(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/metadata-substitution.yaml", config.Name)
	m, _ := manifest.LoadManifestFrom(cxt.Context, config.Name)
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

	before, _ := yaml.Marshal(m.Install[0])
	t.Logf("Before:\n %s", before)
	for _, step := range rm.Install {
		rm.ResolveStep(step)
	}

	s := rm.Install[0]
	after, _ := yaml.Marshal(s)
	t.Logf("After:\n %s", after)

	pms, ok := s.Data["exec"].(map[interface{}]interface{})
	assert.True(t, ok)
	cmd := pms["command"].(string)
	assert.Equal(t, "echo \"name:HELLO version:0.1.0 description:An example Porter configuration image:jeremyrickard/porter-hello-installer:0.1.0\"", cmd)
}

func TestDependencyMetadataAvailableForTemplating(t *testing.T) {
	cxt := context.NewTestContext(t)
	cxt.AddTestFile("testdata/dep-metadata-substitution.yaml", config.Name)

	m, _ := manifest.LoadManifestFrom(cxt.Context, config.Name)
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	rm.bundles = map[string]bundle.Bundle{
		"mysql": {
			Name:        "Azure MySQL",
			Description: "Azure MySQL database as a service",
			Version:     "v1.0.0",
		},
	}

	before, _ := yaml.Marshal(m.Install[0])
	t.Logf("Before:\n %s", before)
	for _, step := range rm.Install {
		rm.ResolveStep(step)
	}

	s := rm.Install[0]
	after, _ := yaml.Marshal(s)
	t.Logf("After:\n %s", after)

	pms, ok := s.Data["exec"].(map[interface{}]interface{})
	assert.True(t, ok)
	cmd := pms["command"].(string)
	assert.Equal(t, "echo \"dep name: Azure MySQL dep version: v1.0.0 dep description: Azure MySQL database as a service\"", cmd)
}

func TestResolveMapParamUnknown(t *testing.T) {
	cxt := context.NewTestContext(t)
	m := &manifest.Manifest{
		Parameters: []manifest.ParameterDefinition{},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

	s := &manifest.Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Parameters": map[string]interface{}{
				"Thing": "{{bundle.parameters.person}}",
			},
		},
	}

	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to render step template Parameters:\n  Thing: '{{bundle.parameters.person}}'\ndescription: a test step\n: Missing variable \"person\"", err.Error())
}

func TestPrepare_fileParam(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/file-param", "/path/to/file")

	m := &manifest.Manifest{
		Parameters: []manifest.ParameterDefinition{
			{
				Name: "file-param",
				Destination: manifest.Location{
					Path: "/path/to/file",
				},
				Schema: definition.Schema{
					Type:    "file",
					Default: "/path/to/file",
				},
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	s := &manifest.Step{
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
	m := &manifest.Manifest{
		Parameters: []manifest.ParameterDefinition{
			{
				Name: "name",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

	s := &manifest.Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.parameters.person }}",
			},
		},
	}

	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to render step template Arguments:\n- '{{ bundle.parameters.person }}'\ndescription: a test step\n: Missing variable \"person\"", err.Error())
}

func TestResolveArray(t *testing.T) {
	os.Setenv("PERSON", "Ralpha")
	defer os.Unsetenv("PERSON")

	cxt := context.NewTestContext(t)
	m := &manifest.Manifest{
		Parameters: []manifest.ParameterDefinition{
			{
				Name: "person",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

	s := &manifest.Step{
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
	m := &manifest.Manifest{
		Parameters: []manifest.ParameterDefinition{
			{
				Name:      "sensitive_param",
				Sensitive: true,
			},
			{
				Name: "regular_param",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

	s := &manifest.Step{
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
	m := &manifest.Manifest{
		Credentials: []manifest.CredentialDefinition{
			{
				Name:     "password",
				Location: manifest.Location{EnvironmentVariable: "PASSWORD"},
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

	s := &manifest.Step{
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

func TestResolveStepOutputs_Install_NoPreexistingClaiml(t *testing.T) {
	cxt := context.NewTestContext(t)

	m := &manifest.Manifest{
		Dependencies: map[string]manifest.Dependency{
			"dep": {
				Tag: "getporter/porter-hello",
			},
		},
	}

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
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

	cxt.FileSystem.WriteFile("/cnab/app/dependencies/dep/outputs/dep_output", []byte("dep_output_value"), 0644)

	s := &manifest.Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
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
	assert.Equal(t, 1, len(args))
	assert.Equal(t, "dep_output_value", args[0].(string))

	// There should now be a sensitive value tracked under the manifest
	assert.Equal(t, []string{"dep_output_value"}, rm.GetSensitiveValues())
}

func TestResolveStepOutputs_fromPreexistingClaim(t *testing.T) {
	cxt := context.NewTestContext(t)

	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Outputs = map[string]interface{}{
		"output": "output_value",
	}

	bytes, err := yaml.Marshal(claim)
	require.NoError(t, err)
	cxt.FileSystem.WriteFile("/cnab/claim.json", bytes, 0644)

	m := &manifest.Manifest{
		Outputs: []manifest.OutputDefinition{
			{
				Name:      "output",
				Sensitive: true,
			},
		},
		Dependencies: map[string]manifest.Dependency{
			"dep": {
				Tag: "getporter/porter-hello",
			},
		},
	}

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionUpgrade, m)
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

	s := &manifest.Step{
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

	err = rm.ResolveStep(s)
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

	m, err := manifest.LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

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

	m, err := manifest.LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

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

func TestResolveMultipleStepOutputsFromPreexistingClaim(t *testing.T) {
	cxt := context.NewTestContext(t)

	databaseURL := "localhost"
	databasePort := "3303"

	// Create a claim to hold the output value for 'output', from the previous action
	claim, err := claim.New("test")
	require.NoError(t, err)

	claim.Outputs = map[string]interface{}{
		"database_url":  databaseURL,
		"database_port": databasePort,
	}

	s := &manifest.Step{
		Data: map[string]interface{}{
			"helm": map[interface{}]interface{}{
				"description": "install wordpress",
				"Arguments": []string{
					"jdbc://{{bundle.outputs.database_url}}:{{bundle.outputs.database_port}}",
				},
			},
		},
	}

	m := &manifest.Manifest{
		Outputs: []manifest.OutputDefinition{
			{
				Name: "database_url",
			},
			{
				Name: "database_port",
			},
		},
		Mixins: []manifest.MixinDeclaration{{Name: "helm"}},
		Install: manifest.Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionUpgrade, m)
	rm.claim = claim

	err = rm.ResolveStep(s)
	require.NoError(t, err)
	helm, ok := s.Data["helm"].(map[interface{}]interface{})
	assert.True(t, ok)
	args, ok := helm["Arguments"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, fmt.Sprintf("jdbc://%s:%s", databaseURL, databasePort), args[0].(string))
}

func TestResolveMissingStepOutputs(t *testing.T) {

	s := &manifest.Step{
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
	m := &manifest.Manifest{
		Mixins: []manifest.MixinDeclaration{{Name: "helm"}},
		Install: manifest.Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to render step template helm:\n  Arguments:\n  - jdbc://{{bundle.outputs.database_url}}:{{bundle.outputs.database_port}}\n  description: install wordpress\n: Missing variable \"database_url\"", err.Error())
}

func TestResolveDependencyParam(t *testing.T) {
	t.Skip("still haven't decided if this is going to be supported")

	s := &manifest.Step{
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
	m := &manifest.Manifest{
		Dependencies: map[string]manifest.Dependency{
			"mysql": {
				Tag: "getporter/porter-mysql",
			},
		},
		Mixins: []manifest.MixinDeclaration{{Name: "helm"}},
		Install: manifest.Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

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

	s := &manifest.Step{
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
	m := &manifest.Manifest{
		Dependencies: map[string]manifest.Dependency{
			"mysql": {
				Tag: "getporter/porter-mysql",
			},
		},
		Mixins: []manifest.MixinDeclaration{{Name: "helm"}},
		Install: manifest.Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

	os.Setenv("DATABASE", "wordpress")
	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to resolve step: unable to render template values: Missing variable \"nope\"", err.Error())
}

func TestManifest_ResolveBundleName(t *testing.T) {
	cxt := context.NewTestContext(t)
	m := &manifest.Manifest{
		Name: "mybundle",
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

	s := &manifest.Step{
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

	wantOutputs := []manifest.OutputDefinition{
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

	m, err := manifest.LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	require.Equal(t, wantOutputs, m.Outputs)
}

func TestReadManifest_Validate_BundleOutput_Error(t *testing.T) {
	cxt := context.NewTestContext(t)

	cxt.AddTestFile("testdata/outputs/bundle-outputs-error.yaml", config.Name)

	_, err := manifest.LoadManifestFrom(cxt.Context, config.Name)
	require.Error(t, err)
}

func TestDependency_Validate(t *testing.T) {
	testcases := []struct {
		name      string
		dep       manifest.Dependency
		wantError string
	}{
		{"version in tag", manifest.Dependency{Tag: "deislabs/azure-mysql:5.7"}, ""},
		{"version ranges", manifest.Dependency{Tag: "deislabs/azure-mysql", Versions: []string{"5.7.x-6"}}, ""},
		{"missing tag", manifest.Dependency{Tag: ""}, "dependency tag is required"},
		{"version double specified", manifest.Dependency{Tag: "deislabs/azure-mysql:5.7", Versions: []string{"5.7.x-6"}}, "dependency tag can only specify REGISTRY/NAME when version ranges are specified"},
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

	cxt.AddTestFile("../manifest/testdata/simple.porter.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)

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

	m, err := manifest.LoadManifestFrom(cxt.Context, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
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
	m := &manifest.Manifest{
		Name: "mybundle",
		ImageMap: map[string]manifest.MappedImage{
			"something": manifest.MappedImage{
				Repository: "blah/blah",
				Digest:     "sha1234:cafebab",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	s := &manifest.Step{
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
	m := &manifest.Manifest{
		Name: "mybundle",
		ImageMap: map[string]manifest.MappedImage{
			"notsomething": manifest.MappedImage{
				Repository: "blah/blah",
				Digest:     "sha1234:cafebab",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	s := &manifest.Step{
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
			reference: "getporter/porter-hello:v0.1.10",
			want: manifest.MappedImage{
				Repository: "getporter/porter-hello",
				Tag:        "v0.1.10",
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
			reference: "getporter/porter-hello:v0.1.0@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want: manifest.MappedImage{
				Repository: "getporter/porter-hello",
				Tag:        "v0.1.0",
				Digest:     "sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			},
		},
	}
	for _, test := range tests {
		got := &manifest.MappedImage{}
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
			reference: "getporter/porter-hello@8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want:      "unable to parse docker image %s: invalid reference format",
		},
		{
			name:      "bad digest",
			reference: "getporter/porter-hello@sha256:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f",
			want:      "unable to parse docker image %s: invalid checksum digest length",
		},
		{
			name:      "bad digest algo",
			reference: "getporter/porter-hello@sha356:8b06c3da72dc9fa7002b9bc1f73a7421b4287c9cf0d3b08633287473707f9a63",
			want:      "unable to parse docker image %s: unsupported digest algorithm",
		},
		{
			name:      "malformed tagged ref",
			reference: "getporter/porter-hello@latest",
			want:      "unable to parse docker image %s: invalid reference format",
		},
		{
			name:      "too many ports tagged ref",
			reference: "deislabs:8080:8080/porter-hello:latest",
			want:      "unable to parse docker image %s: invalid reference format",
		},
	}
	for _, test := range tests {
		got := &manifest.MappedImage{}
		err := resolveImage(got, test.reference)
		assert.EqualError(t, err, fmt.Sprintf(test.want, test.reference))
	}
}

func TestResolveImageWithUpdatedBundle(t *testing.T) {
	cxt := context.NewTestContext(t)
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
	bun := &bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	}

	reloMap := relocation.ImageRelocationMap{}

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "blah/ghost", mi.Repository)
}

func TestResolveImageWithUpdatedMismatchedBundle(t *testing.T) {
	cxt := context.NewTestContext(t)
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
	bun := &bundle.Bundle{
		Images: map[string]bundle.Image{
			"ghost": img,
		},
	}

	reloMap := relocation.ImageRelocationMap{}

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("unable to find image in porter manifest: %s", "ghost"))

}

func TestResolveImageWithRelo(t *testing.T) {
	cxt := context.NewTestContext(t)
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
	bun := &bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	}

	reloMap := relocation.ImageRelocationMap{
		"gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687": "my.registry/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687",
	}

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "my.registry/microservice", mi.Repository)
}

func TestResolveImageRelocationNoMatch(t *testing.T) {
	cxt := context.NewTestContext(t)
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
	bun := &bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	}

	reloMap := relocation.ImageRelocationMap{
		"deislabs/nogood:latest": "cnabio/ghost:latest",
	}

	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.NoError(t, err)
	assert.Equal(t, "deislabs/ghost", rm.ImageMap["machine"].Repository)
}

func TestResolveStepEncoding(t *testing.T) {
	wantValue := `{"test":"value"}`
	os.Setenv("TEST", wantValue)
	defer os.Unsetenv("TEST")

	cxt := context.NewTestContext(t)
	m := &manifest.Manifest{
		Parameters: []manifest.ParameterDefinition{
			{Name: "test"},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionInstall, m)
	s := &manifest.Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Flags": map[string]string{
				"c": "{{bundle.parameters.test}}",
			},
		},
	}

	err := rm.ResolveStep(s)
	require.NoError(t, err)
	flags := s.Data["Flags"].(map[interface{}]interface{})
	assert.Equal(t, flags["c"], wantValue)
}

func TestLoadClaim(t *testing.T) {
	cxt := context.NewTestContext(t)

	s := &manifest.Step{
		Data: map[string]interface{}{
			"helm": map[interface{}]interface{}{
				"description": "install wordpress",
			},
		},
	}
	m := &manifest.Manifest{
		Mixins: []manifest.MixinDeclaration{{Name: "helm"}},
		Install: manifest.Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(cxt.Context, manifest.ActionUpgrade, m)

	// Create a claim and store it in the expected location of the execution environment
	// It holds the output value for 'output', from the previous action
	claim, err := claim.New("test")
	require.NoError(t, err)

	// loadClaim should not error out if the claim does not exist
	err = rm.loadClaim()
	require.NoError(t, err)

	bytes, err := yaml.Marshal(claim)
	require.NoError(t, err)
	cxt.FileSystem.WriteFile("/cnab/claim.json", bytes, 0644)

	err = rm.loadClaim()
	require.NoError(t, err)
}
