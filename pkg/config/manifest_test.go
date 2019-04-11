package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gopkg.in/yaml.v2"
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

func TestLoadManifestWithDependencies(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)
	c.TestContext.AddTestDirectory("testdata/bundles", "bundles")

	require.NoError(t, c.LoadManifest())

	assert.NotNil(t, c.Manifest)
	assert.Equal(t, []string{"helm", "exec"}, c.Manifest.Mixins)
	assert.Len(t, c.Manifest.Install, 2)

	installStep := c.Manifest.Install[0]
	description, _ := installStep.GetDescription()
	assert.NotNil(t, description)

	mixin := installStep.GetMixinName()
	assert.Equal(t, "helm", mixin)
}

func TestConfig_LoadManifest_BundleDependencyNotInstalled(t *testing.T) {
	c := NewTestConfig(t)

	c.TestContext.AddTestFile("testdata/missingdep.porter.yaml", Name)

	err := c.LoadManifest()
	require.Errorf(t, err, "bundle missingdep not installed in PORTER_HOME")
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
	assert.Equal(t, "unable to resolve step: Missing variable \"person\"", err.Error())
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
	assert.Equal(t, "unable to resolve step: Missing variable \"person\"", err.Error())
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
				Name:                "password",
				EnvironmentVariable: "PASSWORD",
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

func TestResolveOutputs(t *testing.T) {
	m := &Manifest{
		outputs: map[string]string{
			"output": "output_value",
		},
		Dependencies: []*Dependency{
			&Dependency{
				Name: "dep",
				m: &Manifest{
					outputs: map[string]string{
						"dep_output": "dep_output_value",
					},
				},
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

func TestResolveMultipleOutputs(t *testing.T) {

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

func TestResolveMissingOutputs(t *testing.T) {

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
	assert.Equal(t, "unable to resolve step: Missing variable \"database_url\"", err.Error())
}

func TestResolveDependencyParam(t *testing.T) {

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
		Dependencies: []*Dependency{
			&Dependency{
				Name: "mysql",
				m: &Manifest{
					Name: "mysql",
					Parameters: []ParameterDefinition{
						{
							Name: "database",
						},
					},
				},
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
		Dependencies: []*Dependency{
			&Dependency{
				Name: "mysql",
				m: &Manifest{
					Name: "mysql",
					Parameters: []ParameterDefinition{
						{
							Name: "database",
						},
					},
				},
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
	assert.Equal(t, "unable to resolve step: Missing variable \"nope\"", err.Error())
}

func TestDependency_Validate_NameRequired(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)
	c.TestContext.AddTestDirectory("testdata/bundles", "bundles")

	err := c.LoadManifest()
	require.NoError(t, err)

	// Sabotage!
	c.Manifest.Dependencies[0].Name = ""

	err = c.Manifest.Dependencies[0].Validate()
	assert.EqualError(t, err, "dependency name is required")
}

func TestManifest_MergeDependency(t *testing.T) {
	m := &Manifest{
		Mixins: []string{"helm"},
		Install: Steps{
			&Step{Data: map[string]interface{}{"helm": map[interface{}]interface{}{"description": "install wordpress"}}},
		},
		Upgrade: Steps{
			&Step{Data: map[string]interface{}{"helm": map[interface{}]interface{}{"description": "upgrade wordpress"}}},
		},
		Uninstall: Steps{
			&Step{Data: map[string]interface{}{"helm": map[interface{}]interface{}{"description": "uninstall wordpress"}}},
		},
	}

	dep := &Dependency{
		m: &Manifest{
			Mixins: []string{"exec", "helm"},
			Install: Steps{
				&Step{Data: map[string]interface{}{"helm": map[interface{}]interface{}{"description": "install mysql"}}},
			},
			Upgrade: Steps{
				&Step{Data: map[string]interface{}{"helm": map[interface{}]interface{}{"description": "upgrade mysql"}}},
			},
			Uninstall: Steps{
				&Step{Data: map[string]interface{}{"helm": map[interface{}]interface{}{"description": "uninstall mysql"}}},
			},
			Credentials: []CredentialDefinition{
				{Name: "kubeconfig", Path: "/root/.kube/config"},
			},
		},
	}

	err := m.MergeDependency(dep)
	require.NoError(t, err)

	assert.Equal(t, []string{"exec", "helm"}, m.Mixins)

	assert.Len(t, m.Install, 2)
	description, _ := m.Install[0].GetDescription()
	assert.Equal(t, "install mysql", description)
	description, _ = m.Install[1].GetDescription()
	assert.Equal(t, "install wordpress", description)

	assert.Len(t, m.Upgrade, 2)
	description, _ = m.Upgrade[0].GetDescription()
	assert.Equal(t, "upgrade mysql", description)
	description, _ = m.Upgrade[1].GetDescription()
	assert.Equal(t, "upgrade wordpress", description)

	assert.Len(t, m.Uninstall, 2)
	description, _ = m.Uninstall[0].GetDescription()
	assert.Equal(t, "uninstall wordpress", description)
	description, _ = m.Uninstall[1].GetDescription()
	assert.Equal(t, "uninstall mysql", description)

	assert.Len(t, m.Credentials, 1)
}

func TestMergeCredentials(t *testing.T) {
	testcases := []struct {
		name               string
		c1, c2, wantResult CredentialDefinition
		wantError          string
	}{
		{
			name:       "combine path and environment variable",
			c1:         CredentialDefinition{Name: "foo", Path: "p1"},
			c2:         CredentialDefinition{Name: "foo", EnvironmentVariable: "v2"},
			wantResult: CredentialDefinition{Name: "foo", Path: "p1", EnvironmentVariable: "v2"},
		},
		{
			name:       "same path",
			c1:         CredentialDefinition{Name: "foo", Path: "p"},
			c2:         CredentialDefinition{Name: "foo", Path: "p"},
			wantResult: CredentialDefinition{Name: "foo", Path: "p"},
		},
		{
			name:      "conflicting path",
			c1:        CredentialDefinition{Name: "foo", Path: "p1"},
			c2:        CredentialDefinition{Name: "foo", Path: "p2"},
			wantError: "cannot merge credential foo: conflict on path",
		},
		{
			name:       "same environment variable",
			c1:         CredentialDefinition{Name: "foo", EnvironmentVariable: "v"},
			c2:         CredentialDefinition{Name: "foo", EnvironmentVariable: "v"},
			wantResult: CredentialDefinition{Name: "foo", EnvironmentVariable: "v"},
		},
		{
			name:      "conflicting environment variable",
			c1:        CredentialDefinition{Name: "foo", EnvironmentVariable: "v1"},
			c2:        CredentialDefinition{Name: "foo", EnvironmentVariable: "v2"},
			wantError: "cannot merge credential foo: conflict on environment variable",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := mergeCredentials(tc.c1, tc.c2)

			if tc.wantError == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantResult, result)
			} else {
				require.Contains(t, err.Error(), tc.wantError)
			}
		})
	}

}

func TestManifest_ApplyBundleOutputs(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/simple.porter.yaml", Name)

	require.NoError(t, c.LoadManifest())

	depStep := c.Manifest.Install[0]
	err := c.Manifest.ApplyOutputs(depStep, []string{"foo=bar"})
	require.NoError(t, err)

	assert.Contains(t, c.Manifest.outputs, "foo")
	assert.Equal(t, "bar", c.Manifest.outputs["foo"])
}

func TestManifest_ApplyDependencyOutputs(t *testing.T) {
	testcases := []struct {
		name        string
		rawOutputs  []string
		wantOutputs map[string]string
		wantError   string
	}{
		{
			name:        "happy path",
			rawOutputs:  []string{"host=localhost"},
			wantOutputs: map[string]string{"host": "localhost"},
		},
		{
			name:        "value with equals sign",
			rawOutputs:  []string{"cert=abc123==="},
			wantOutputs: map[string]string{"cert": "abc123==="},
		},
		{
			name:       "missing equals sign",
			rawOutputs: []string{"foo"},
			wantError:  "invalid output assignment",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewTestConfig(t)
			c.SetupPorterHome()

			c.TestContext.AddTestFile("testdata/porter.yaml", Name)
			c.TestContext.AddTestDirectory("testdata/bundles", "bundles")

			require.NoError(t, c.LoadManifest())

			depStep := c.Manifest.Install[0]
			err := c.Manifest.ApplyOutputs(depStep, tc.rawOutputs)
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, err.Error(), tc.wantError)
				return
			}

			depM := c.Manifest.Dependencies[0].m
			for wantKey, wantValue := range tc.wantOutputs {
				assert.Contains(t, depM.outputs, wantKey)
				assert.Equal(t, wantValue, depM.outputs[wantKey])
			}
		})
	}
}

func TestManifest_MergeParameters(t *testing.T) {
	dep := &Dependency{
		Name:       "mysql",
		Parameters: map[string]string{"database": "wordpress"},
		m: &Manifest{
			Name: "mysql",
			Parameters: []ParameterDefinition{
				{Name: "database"},
			},
		},
	}
	m := &Manifest{
		Name:         "wordpress",
		Dependencies: []*Dependency{dep},
	}

	err := m.MergeParameters(dep)
	require.NoError(t, err)

	require.Len(t, m.Parameters, 1)
	assert.Equal(t, "wordpress", m.Parameters[0].DefaultValue)
}
