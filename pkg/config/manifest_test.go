package config

import (
	"io/ioutil"
	"testing"

	"github.com/deislabs/cnab-go/bundle/definition"
	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadManifest_URL(t *testing.T) {
	c := NewTestConfig(t)
	url := "https://raw.githubusercontent.com/deislabs/porter/master/pkg/config/testdata/simple.porter.yaml"
	m, err := c.ReadManifest(url)

	require.NoError(t, err)
	assert.Equal(t, "hello", m.Name)
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
}

func TestSetDefaultInvocationImage(t *testing.T) {
	c := NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/missing-invocation-image.porter.yaml", Name)
	m, err := c.ReadManifest(Name)
	require.NoError(t, err)
	assert.Equal(t, "deislabs/missing-invocation-image-installer:" + m.Version, m.Image)
}

func TestReadManifest_Validate_MissingFile(t *testing.T) {
	c := NewTestConfig(t)
	_, err := c.ReadManifest("fake-porter.yaml")

	assert.EqualError(t, err, "the specified porter configuration file fake-porter.yaml does not exist")
}

func TestMixinDeclaration_UnmarshalYAML(t *testing.T) {
	c := NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/mixin-with-config.yaml", Name)
	m, err := c.ReadManifest(Name)

	require.NoError(t, err)
	assert.Len(t, m.Mixins, 2, "expected 2 mixins")
	assert.Equal(t, "exec", m.Mixins[0].Name)
	assert.Equal(t, "az", m.Mixins[1].Name)
	assert.Equal(t, map[interface{}]interface{}{"extensions": []interface{}{"iot"}}, m.Mixins[1].Config)
}

func TestMixinDeclaration_UnmarshalYAML_Invalid(t *testing.T) {
	c := NewTestConfig(t)
	c.TestContext.AddTestFile("testdata/mixin-with-bad-config.yaml", Name)
	_, err := c.ReadManifest(Name)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "mixin declaration contained more than one mixin")
}

func TestMixinDeclaration_MarshalYAML(t *testing.T) {
	m := struct {
		Mixins []MixinDeclaration
	}{
		[]MixinDeclaration{
			{Name: "exec"},
			{Name: "az", Config: map[interface{}]interface{}{"extensions": []interface{}{"iot"}}},
		},
	}

	gotYaml, err := yaml.Marshal(m)
	require.NoError(t, err, "could not marshal data")

	wantYaml, err := ioutil.ReadFile("testdata/mixin-with-config.yaml")
	require.NoError(t, err, "could not read testdata")

	assert.Equal(t, string(wantYaml), string(gotYaml))
}

func TestValidateParameterDefinition(t *testing.T) {
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
