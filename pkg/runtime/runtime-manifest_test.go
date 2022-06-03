package runtime

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/yaml"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-to-oci/relocation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveMapParam(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.Setenv("PERSON", "Ralpha")

	m := &manifest.Manifest{
		Parameters: manifest.ParameterDefinitions{
			"person": {
				Name: "person",
			},
			"place": {
				Name:    "place",
				ApplyTo: []string{cnab.ActionInstall},
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
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
	pms, ok := s.Data["Parameters"].(map[string]interface{})
	assert.True(t, ok)
	val, ok := pms["Thing"].(string)
	assert.True(t, ok)
	assert.Equal(t, "Ralpha", val)
	assert.NotContains(t, "place", pms, "parameters that don't apply to the current action should not be resolved")

	err = rm.Initialize()
	assert.NoError(t, err)
}

func TestResolvePathParam(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		Parameters: manifest.ParameterDefinitions{
			"person": {
				Name: "person",
				Destination: manifest.Location{
					Path: "person.txt",
				},
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
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
	pms, ok := s.Data["Parameters"].(map[string]interface{})
	assert.True(t, ok)
	val, ok := pms["Thing"].(string)
	assert.True(t, ok)
	assert.Equal(t, "person.txt", val)
}

func TestMetadataAvailableForTemplating(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/metadata-substitution.yaml", config.Name)
	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "LoadManifestFrom")
	rm := NewRuntimeManifest(c.Context, cnab.ActionInstall, m)

	before, _ := yaml.Marshal(m.Install[0])
	t.Logf("Before:\n %s", before)
	for _, step := range rm.Install {
		err := rm.ResolveStep(step)
		require.NoError(t, err)
	}

	s := rm.Install[0]
	after, _ := yaml.Marshal(s)
	t.Logf("After:\n %s", after)

	pms, ok := s.Data["exec"].(map[string]interface{})
	assert.True(t, ok)
	cmd := pms["command"].(string)
	assert.Equal(t, "echo \"name:porter-hello version:0.1.0 description:An example Porter configuration image:jeremyrickard/porter-hello:39a022ca907e26c3d8fffabd4bb8dbbc\"", cmd)
}

func TestDependencyMetadataAvailableForTemplating(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/dep-metadata-substitution.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "LoadManifestFrom failed")
	rm := NewRuntimeManifest(c.Context, cnab.ActionInstall, m)
	rm.bundles = map[string]cnab.ExtendedBundle{
		"mysql": {bundle.Bundle{
			Name:        "Azure MySQL",
			Description: "Azure MySQL database as a service",
			Version:     "v1.0.0",
		}},
	}

	before, _ := yaml.Marshal(m.Install[0])
	t.Logf("Before:\n %s", before)
	for _, step := range rm.Install {
		rm.ResolveStep(step)
	}

	s := rm.Install[0]
	after, _ := yaml.Marshal(s)
	t.Logf("After:\n %s", after)

	pms, ok := s.Data["exec"].(map[string]interface{})
	assert.True(t, ok)
	cmd := pms["command"].(string)
	assert.Equal(t, "echo \"dep name: Azure MySQL dep version: v1.0.0 dep description: Azure MySQL database as a service\"", cmd)
}

func TestResolveMapParamUnknown(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		Parameters: manifest.ParameterDefinitions{},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

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

func TestResolveArrayUnknown(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		Parameters: manifest.ParameterDefinitions{
			"name": {
				Name: "name",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

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
	assert.Contains(t, err.Error(), `Missing variable "person"`)
}

func TestResolveArray(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.Setenv("PERSON", "Ralpha")
	m := &manifest.Manifest{
		Parameters: manifest.ParameterDefinitions{
			"person": {
				Name: "person",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

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
	cxt := portercontext.NewTestContext(t)
	cxt.Setenv("SENSITIVE_PARAM", "deliciou$dubonnet")
	cxt.Setenv("REGULAR_PARAM", "regular param value")

	m := &manifest.Manifest{
		Parameters: manifest.ParameterDefinitions{
			"sensitive_param": {
				Name:      "sensitive_param",
				Sensitive: true,
			},
			"regular_param": {
				Name: "regular_param",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

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
	cxt := portercontext.NewTestContext(t)
	cxt.Setenv("PASSWORD", "deliciou$dubonnet")

	m := &manifest.Manifest{
		Credentials: manifest.CredentialDefinitions{
			"password": {
				Name:     "password",
				Location: manifest.Location{EnvironmentVariable: "PASSWORD"},
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

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

func TestResolveStep_DependencyOutput(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.Setenv("PORTER_MYSQL_PASSWORD_DEP_OUTPUT", "password")
	cxt.Setenv("PORTER_MYSQL_ROOT_PASSWORD_DEP_OUTPUT", "mysql-password")

	m := &manifest.Manifest{
		Dependencies: manifest.Dependencies{
			RequiredDependencies: []*manifest.RequiredDependency{
				{
					Name:   "mysql",
					Bundle: manifest.BundleCriteria{Reference: "getporter/porter-mysql"},
				},
			},
		},
		TemplateVariables: []string{
			"bundle.dependencies.mysql.outputs.password",
			"bundle.dependencies.mysql.outputs.root-password",
		},
	}

	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
	ps := cnab.ParameterSources{}
	ps.SetParameterFromDependencyOutput("porter-mysql-password", "mysql", "password")
	ps.SetParameterFromDependencyOutput("porter-mysql-root-password", "mysql", "root-password")
	rm.bundle = cnab.ExtendedBundle{bundle.Bundle{
		Custom: map[string]interface{}{
			cnab.ParameterSourcesExtensionKey: ps,
		},
		RequiredExtensions: []string{cnab.ParameterSourcesExtensionKey},
	}}

	rm.bundles = map[string]cnab.ExtendedBundle{
		"mysql": {bundle.Bundle{
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
		}},
	}

	s := &manifest.Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.dependencies.mysql.outputs.password }}",
				"{{ bundle.dependencies.mysql.outputs.root-password }}",
			},
		},
	}

	// Prior to resolving step values, this method should return an empty string array
	assert.Equal(t, rm.GetSensitiveValues(), []string{})

	err := rm.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, []interface{}{"password", "mysql-password"}, args, "Incorrect template args passed to the mixin step")

	// There should now be a sensitive value tracked under the manifest
	gotSensitiveValues := rm.GetSensitiveValues()
	sort.Strings(gotSensitiveValues)
	assert.Equal(t, []string{"mysql-password", "password"}, gotSensitiveValues, "Incorrect values were marked as sensitive")
}

func TestResolveInMainDict(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/param-test-in-block.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := NewRuntimeManifest(c.Context, cnab.ActionInstall, m)

	installStep := rm.Install[0]

	rm.Setenv("COMMAND", "echo hello world")
	err = rm.ResolveStep(installStep)
	assert.NoError(t, err)

	assert.NotNil(t, installStep.Data)
	t.Logf("install data %v", installStep.Data)
	exec := installStep.Data["exec"].(map[string]interface{})
	assert.NotNil(t, exec)
	command := exec["command"].(interface{})
	assert.NotNil(t, command)
	cmdVal, ok := command.(string)
	assert.True(t, ok)
	assert.Equal(t, "echo hello world", cmdVal)
}

func TestResolveSliceWithAMap(t *testing.T) {
	c := config.NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/slice-test.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := NewRuntimeManifest(c.Context, cnab.ActionInstall, m)

	installStep := rm.Install[0]

	rm.Setenv("COMMAND", "echo hello world")
	err = rm.ResolveStep(installStep)
	assert.NoError(t, err)

	assert.NotNil(t, installStep.Data)
	t.Logf("install data %v", installStep.Data)
	exec := installStep.Data["exec"].(map[string]interface{})
	assert.NotNil(t, exec)
	flags := exec["flags"].(map[string]interface{})
	assert.Len(t, flags, 1)
	assert.Equal(t, "echo hello world", flags["c"].(string))
	assert.NotNil(t, flags)
}

func TestResolveMissingStepOutputs(t *testing.T) {

	s := &manifest.Step{
		Data: map[string]interface{}{
			"helm": map[string]interface{}{
				"description": "install wordpress",
				"Arguments": []string{
					"jdbc://{{bundle.outputs.database_url}}:{{bundle.outputs.database_port}}",
				},
			},
		},
	}

	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		Mixins: []manifest.MixinDeclaration{{Name: "helm"}},
		Install: manifest.Steps{
			s,
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

	err := rm.ResolveStep(s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `Missing variable "database_url"`)
}

func TestResolveSensitiveOutputs(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		Outputs: manifest.OutputDefinitions{
			"username": {
				Name: "username",
			},
			"password": {
				Name:      "password",
				Sensitive: true,
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
	rm.outputs = map[string]string{
		"username": "sally",
		"password": "top$ecret!",
	}

	s := &manifest.Step{
		Data: map[string]interface{}{
			"description": "a test step",
			"Arguments": []string{
				"{{ bundle.outputs.username }}",
				"{{ bundle.outputs.password }}",
			},
		},
	}

	err := rm.ResolveStep(s)
	require.NoError(t, err)

	args, ok := s.Data["Arguments"].([]interface{})
	require.True(t, ok)
	require.Equal(t, 2, len(args))
	require.Equal(t, "sally", args[0])
	require.Equal(t, "top$ecret!", args[1])

	// There should be only one sensitive value being tracked
	require.Equal(t, []string{"top$ecret!"}, rm.GetSensitiveValues())
}

func TestManifest_ResolveBundleName(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		Name: "mybundle",
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

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

func TestDependency_Validate(t *testing.T) {
	testcases := []struct {
		name       string
		dep        manifest.RequiredDependency
		wantOutput string
		wantError  string
	}{
		{
			name:       "version in reference",
			dep:        manifest.RequiredDependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql:5.7"}},
			wantOutput: "",
			wantError:  "",
		}, {
			name:       "version ranges",
			dep:        manifest.RequiredDependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql", Version: "5.7.x-6"}},
			wantOutput: "",
			wantError:  "",
		}, {
			name:       "missing reference",
			dep:        manifest.RequiredDependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: ""}},
			wantOutput: "",
			wantError:  `reference is required for dependency "mysql"`,
		}, {
			name:       "version double specified",
			dep:        manifest.RequiredDependency{Name: "mysql", Bundle: manifest.BundleCriteria{Reference: "deislabs/azure-mysql:5.7", Version: "5.7.x-6"}},
			wantOutput: "",
			wantError:  `reference for dependency "mysql" can only specify REGISTRY/NAME when version ranges are specified`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			cxt := portercontext.NewTestContext(t)

			err := tc.dep.Validate(cxt.Context)

			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.Equal(t, tc.wantError, err.Error())
			}

			gotOutput := cxt.GetOutput()
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

	rm := NewRuntimeManifest(c.Context, cnab.ActionInstall, m)

	err = rm.ApplyStepOutputs(map[string]string{"name": "world"})
	require.NoError(t, err)

	assert.Contains(t, rm.outputs, "name")
	assert.Equal(t, "world", rm.outputs["name"])
}

func makeBoolPtr(value bool) *bool {
	return &value
}

func TestManifest_ResolveImageMap(t *testing.T) {
	c := config.NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/porter-images.yaml", config.Name)

	m, err := manifest.LoadManifestFrom(context.Background(), c.Config, config.Name)
	require.NoError(t, err, "could not load manifest")

	rm := NewRuntimeManifest(c.Context, cnab.ActionInstall, m)
	expectedImage, ok := m.ImageMap["something"]
	require.True(t, ok, "couldn't get expected image")
	expectedRef := fmt.Sprintf("%s@%s", expectedImage.Repository, expectedImage.Digest)
	step := rm.Install[0]
	err = rm.ResolveStep(step)
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

	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		Name: "mybundle",
		ImageMap: map[string]manifest.MappedImage{
			"something": manifest.MappedImage{
				Repository: "blah/blah",
				Digest:     "sha1234:cafebab",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
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

	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		Name: "mybundle",
		ImageMap: map[string]manifest.MappedImage{
			"notsomething": manifest.MappedImage{
				Repository: "blah/blah",
				Digest:     "sha1234:cafebab",
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
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
			assert.NoError(t, err)
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
	cxt := portercontext.NewTestContext(t)
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
	bun := cnab.ExtendedBundle{bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	}}

	reloMap := relocation.ImageRelocationMap{}

	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "blah/ghost", mi.Repository)
}

func TestResolveImageWithUpdatedMismatchedBundle(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
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
	bun := cnab.ExtendedBundle{bundle.Bundle{
		Images: map[string]bundle.Image{
			"ghost": img,
		},
	}}

	reloMap := relocation.ImageRelocationMap{}

	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("unable to find image in porter manifest: %s", "ghost"))

}

func TestResolveImageWithRelo(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
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
	bun := cnab.ExtendedBundle{bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	}}

	reloMap := relocation.ImageRelocationMap{
		"gabrtv/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687": "my.registry/microservice@sha256:cca460afa270d4c527981ef9ca4989346c56cf9b20217dcea37df1ece8120687",
	}

	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.NoError(t, err)
	mi := rm.ImageMap["machine"]
	assert.Equal(t, "my.registry/microservice", mi.Repository)
}

func TestResolveImageRelocationNoMatch(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
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
	bun := cnab.ExtendedBundle{bundle.Bundle{
		Images: map[string]bundle.Image{
			"machine": img,
		},
	}}

	reloMap := relocation.ImageRelocationMap{
		"deislabs/nogood:latest": "cnabio/ghost:latest",
	}

	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
	err := rm.ResolveImages(bun, reloMap)
	assert.NoError(t, err)
	assert.Equal(t, "deislabs/ghost", rm.ImageMap["machine"].Repository)
}

func TestResolveStepEncoding(t *testing.T) {
	cxt := portercontext.NewTestContext(t)

	wantValue := `{"test":"value"}`
	cxt.Setenv("TEST", wantValue)

	m := &manifest.Manifest{
		Parameters: manifest.ParameterDefinitions{
			"test": {Name: "test"},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)
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
	flags := s.Data["Flags"].(map[string]interface{})
	assert.Equal(t, flags["c"], wantValue)
}

func TestResolveInstallation(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	cxt.Setenv(config.EnvPorterInstallationNamespace, "mynamespace")
	cxt.Setenv(config.EnvPorterInstallationName, "mybun")

	m := &manifest.Manifest{}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

	s := &manifest.Step{
		Data: map[string]interface{}{
			"description": "Do a helm release",
			"ns":          "{{ installation.namespace }}",
			"release":     "{{ installation.name }}",
		},
	}

	err := rm.ResolveStep(s)
	require.NoError(t, err, "ResolveStep failed")

	assert.Equal(t, "mynamespace", s.Data["ns"], "installation.namespace was not rendered")
	assert.Equal(t, "mybun", s.Data["release"], "installation.name was not rendered")
}

func TestResolveCustomMetadata(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{
		Custom: map[string]interface{}{
			"foo": "foobar",
			"myApp": map[string]interface{}{
				"featureFlags": map[string]bool{
					"featureA": true,
				},
			},
		},
	}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

	s := &manifest.Step{
		Data: map[string]interface{}{
			"description": "Do a helm release",
			"release":     "{{ bundle.custom.foo }}",
			"featureA":    "{{ bundle.custom.myApp.featureFlags.featureA }}",
		},
	}

	err := rm.ResolveStep(s)
	require.NoError(t, err, "ResolveStep failed")

	assert.Equal(t, "foobar", s.Data["release"], "custom metadata was not rendered")
	assert.Equal(t, "true", s.Data["featureA"], "nested custom metadata was not rendered")
}

func TestResolveEnvironmentVariable(t *testing.T) {
	cxt := portercontext.NewTestContext(t)
	m := &manifest.Manifest{}
	rm := NewRuntimeManifest(cxt.Context, cnab.ActionInstall, m)

	s := &manifest.Step{
		Data: map[string]interface{}{
			"description": "Read an environment variable",
			"someInput":   "{{ env.foo }}",
			"moreInput":   "{{ env.BAR }}",
		},
	}

	cxt.Setenv("foo", "foo-value")
	cxt.Setenv("BAR", "bar-value")
	err := rm.ResolveStep(s)
	require.NoError(t, err, "ResolveStep failed")

	assert.Equal(t, "foo-value", s.Data["someInput"], "expected lower-case foo env var was resolved")
	assert.Equal(t, "bar-value", s.Data["moreInput"], "expected upper-case BAR env var was resolved")
}
