package porter

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPublish_PorterYamlExists(t *testing.T) {

	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	pwd, err := os.Getwd()
	require.NoError(t, err, "should not have gotten an error")
	t.Log(p.TestConfig.TestContext.FindBinDir())
	p.TestConfig.TestContext.AddTestDirectory("testdata", pwd)
	opts := PublishOptions{}
	err = opts.Validate(p.Porter)
	require.NoError(t, err, "should have no error")
}

func TestPublish_PorterYamlDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	opts := PublishOptions{}
	err := opts.Validate(p.Porter)
	require.Error(t, err, "should have no error")
}

func TestPublish_ValidTag(t *testing.T) {

	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	pwd, err := os.Getwd()
	require.NoError(t, err, "should not have gotten an error")
	p.TestConfig.TestContext.AddTestDirectory("testdata", pwd)
	opts := PublishOptions{}
	opts.Tag = "somerepo/thing:10"

	err = opts.Validate(p.Porter)
	require.NoError(t, err, "should have no error")
}

func TestPublish_InValidTag(t *testing.T) {

	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	pwd, err := os.Getwd()
	require.NoError(t, err, "should not have gotten an error")
	p.TestConfig.TestContext.AddTestDirectory("testdata", pwd)
	opts := PublishOptions{}
	opts.Tag = "someinvalid/repo/thing:10:10"

	err = opts.Validate(p.Porter)
	require.Error(t, err, "should have no error")
}
