package config

import (
	"testing"

	"github.com/deislabs/cnab-go/bundle/definition"
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

func TestReadManifest_Validate_MissingFile(t *testing.T) {
	c := NewTestConfig(t)
	_, err := c.ReadManifest("fake-porter.yaml")

	assert.EqualError(t, err, "the specified porter configuration file fake-porter.yaml does not exist")
}
