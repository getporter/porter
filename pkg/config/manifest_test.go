package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadManifest(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)

	require.NoError(t, c.LoadManifest(Name))

	assert.NotNil(t, c.Manifest)
	assert.Equal(t, []string{"exec"}, c.Manifest.Mixins)
	assert.Len(t, c.Manifest.Install, 1)

	installStep := c.Manifest.Install[0]
	assert.NotNil(t, installStep.Description)

	mixin := installStep.GetMixinName()
	assert.Equal(t, "exec", mixin)

	data := installStep.GetMixinData()
	wantData := `arguments:
- -c
- Hello World!
command: bash
`
	assert.Equal(t, wantData, data)
}

func TestManifest_Validate(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)

	err := c.LoadManifest(Name)
	require.NoError(t, err)

	assert.NoError(t, c.Manifest.Validate())
}

func TestAction_Validate_RequireMixinDeclaration(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)

	err := c.LoadManifest(Name)
	require.NoError(t, err)

	// Sabotage!
	c.Manifest.Mixins = []string{}

	err = c.Manifest.Install.Validate(c.Manifest)
	assert.EqualError(t, err, "mixin (exec) was not declared")
}

func TestAction_Validate_RequireMixinData(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)

	err := c.LoadManifest(Name)
	require.NoError(t, err)

	// Sabotage!
	c.Manifest.Install[0].Data = nil

	err = c.Manifest.Install.Validate(c.Manifest)
	assert.EqualError(t, err, "no mixin specified")
}

func TestAction_Validate_RequireSingleMixinData(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)

	err := c.LoadManifest(Name)
	require.NoError(t, err)

	// Sabotage!
	c.Manifest.Install[0].Data["rando-mixin"] = ""

	err = c.Manifest.Install.Validate(c.Manifest)
	assert.EqualError(t, err, "more than one mixin specified")
}

func TestResolveMapParam(t *testing.T) {
	m := &Manifest{
		Parameters: []ParameterDefinition{
			ParameterDefinition{
				Name: "person",
			},
		},
	}

	os.Setenv("PERSON", "Ralpha")
	s := &Step{
		Description: "a test step",
		Data: map[string]interface{}{
			"Parameters": map[string]interface{}{
				"Thing": map[string]interface{}{
					"source": "bundle.parameters.person",
				},
			},
		},
	}

	err := m.ResolveStep(s)
	require.NoError(t, err)
	pms, ok := s.Data["Parameters"].(map[string]interface{})
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
		Description: "a test step",
		Data: map[string]interface{}{
			"Parameters": map[string]interface{}{
				"Thing": map[string]interface{}{
					"source": "bundle.parameters.person",
				},
			},
		},
	}

	err := m.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to set value for Thing: no value found for source specification: bundle.parameters.person", err.Error())
}

func TestResolveArrayUnknown(t *testing.T) {
	m := &Manifest{
		Parameters: []ParameterDefinition{
			ParameterDefinition{
				Name: "name",
			},
		},
	}

	s := &Step{
		Description: "a test step",
		Data: map[string]interface{}{
			"Arguments": []string{
				"source: bundle.parameters.person",
			},
		},
	}

	err := m.ResolveStep(s)
	require.Error(t, err)
	assert.Equal(t, "unable to source value: no value found for source specification: bundle.parameters.person", err.Error())
}

func TestResolveArray(t *testing.T) {
	m := &Manifest{
		Parameters: []ParameterDefinition{
			ParameterDefinition{
				Name: "person",
			},
		},
	}

	os.Setenv("PERSON", "Ralpha")
	s := &Step{
		Description: "a test step",
		Data: map[string]interface{}{
			"Arguments": []string{
				"source: bundle.parameters.person",
			},
		},
	}

	err := m.ResolveStep(s)
	require.NoError(t, err)
	args, ok := s.Data["Arguments"].([]string)
	assert.True(t, ok)
	assert.Equal(t, "Ralpha", args[0])
}

func TestDependency_Validate_NameRequired(t *testing.T) {
	c := NewTestConfig(t)
	c.SetupPorterHome()

	c.TestContext.AddTestFile("testdata/porter.yaml", Name)

	err := c.LoadManifest(Name)
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
			&Step{Description: "install wordpress"},
		},
		Uninstall: Steps{
			&Step{Description: "uninstall wordpress"},
		},
	}

	depM := &Manifest{
		Mixins: []string{"exec", "helm"},
		Install: Steps{
			&Step{Description: "install mysql"},
		},
		Uninstall: Steps{
			&Step{Description: "uninstall mysql"},
		},
		Credentials: []CredentialDefinition{
			{Name: "kubeconfig", Path: "/root/.kube/config"},
		},
	}

	err := m.MergeDependency(depM)
	require.NoError(t, err)

	assert.Equal(t, []string{"exec", "helm"}, m.Mixins)

	assert.Len(t, m.Install, 2)
	assert.Equal(t, "install mysql", m.Install[0].Description)
	assert.Equal(t, "install wordpress", m.Install[1].Description)

	assert.Len(t, m.Uninstall, 2)
	assert.Equal(t, "uninstall wordpress", m.Uninstall[0].Description)
	assert.Equal(t, "uninstall mysql", m.Uninstall[1].Description)

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
