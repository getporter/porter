package porter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish_Validate_PorterYamlExists(t *testing.T) {

	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	pwd, err := os.Getwd()
	require.NoError(t, err, "should not have gotten an error obtaining current working directory")

	p.TestConfig.TestContext.AddTestFile("testdata/porter.yaml", filepath.Join(pwd, "porter.yaml"))
	opts := PublishOptions{}
	err = opts.Validate(p.Context)
	require.NoError(t, err, "validating should not have failed")

}

func TestPublish_Validate_PorterYamlDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	opts := PublishOptions{}
	err := opts.Validate(p.Context)
	require.Error(t, err, "validation should have failed")
	assert.EqualError(
		t,
		err,
		"could not find porter.yaml in the current directory, make sure you are in the right directory or specify the porter manifest with --file",
		"porter.yaml not present so should have failed validation",
	)
}

func TestPublish_Validate_ArchivePath(t *testing.T) {
	p := NewTestPorter(t)
	p.TestConfig.SetupPorterHome()
	pwd, err := p.GetHomeDir()
	require.NoError(t, err, "should not have gotten an error obtaining current working directory")

	opts := PublishOptions{
		ArchiveFile: filepath.Join(pwd, "mybuns.tgz"),
	}
	err = opts.Validate(p.Context)
	assert.EqualError(t, err, "unable to access --archive /root/.porter/mybuns.tgz: open /root/.porter/mybuns.tgz: file does not exist")

	p.FileSystem.WriteFile(filepath.Join(pwd, "mybuns.tgz"), []byte("mybuns"), os.ModePerm)
	err = opts.Validate(p.Context)
	assert.EqualError(t, err, "must provide a value for --tag of the form REGISTRY/bundle:tag")

	opts.Tag = "myreg/mybuns:v0.1.0"
	err = opts.Validate(p.Context)
	require.NoError(t, err, "validating should not have failed")
}
