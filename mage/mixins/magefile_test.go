package mixins

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstall(t *testing.T) {
	magefile := NewMagefile("github.com/mymixin/test-mixin", "testmixin", "testdata/bin/mixins/testmixin")

	// Change the porter home to a safe place for the test to write to
	require.NoError(t, os.MkdirAll("testdata/porter_home", 0700))
	os.Setenv("PORTER_HOME", "testdata/porter_home")
	defer os.Unsetenv("PORTER_HOME")

	magefile.Install()

	assert.DirExists(t, "testdata/porter_home/mixins/testmixin", "The mixin directory doesn't exist")
	assert.FileExists(t, "testdata/porter_home/mixins/testmixin/testmixin", "The client wasn't installed")
	assert.DirExists(t, "testdata/porter_home/mixins/testmixin/runtimes", "The mixin runtimes directory doesn't exist")
	assert.FileExists(t, "testdata/porter_home/mixins/testmixin/runtimes/testmixin-runtime", "The runtime wasn't installed")
}
