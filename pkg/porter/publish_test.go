package porter

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish_PorterYamlExists(t *testing.T) {

	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	pwd, err := os.Getwd()
	require.NoError(t, err, "should not have gotten an error obtaining current working directory")
	t.Log(p.TestConfig.TestContext.FindBinDir())
	p.TestConfig.TestContext.AddTestDirectory("testdata", pwd)
	opts := PublishOptions{}
	err = opts.Validate(p.Porter)
	require.NoError(t, err, "validating should not have failed")

}

func TestPublish_PorterYamlDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	opts := PublishOptions{}
	err := opts.Validate(p.Porter)
	require.Error(t, err, "validation should have failed")
	assert.EqualError(
		t,
		err,
		"could not find porter.yaml. run `porter create` and `porter build` to create a new bundle before publishing",
		"porter.yaml not present so should have failed validation",
	)
}
