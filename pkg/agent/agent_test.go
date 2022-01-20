// +build integration

package agent

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecute(t *testing.T) {
	home := makeTestPorterHome(t)
	defer os.RemoveAll(home)
	cfg := "testdata"

	stdoutBuff := &bytes.Buffer{}
	stderrBuff := &bytes.Buffer{}
	stdout = stdoutBuff
	stderr = stderrBuff

	err, run := Execute([]string{"help"}, home, cfg)
	require.NoError(t, err)
	assert.True(t, run, "porter should have run")
	gotStderr := stderrBuff.String()
	assert.Contains(t, gotStderr, "porter version", "the agent should always print the porter CLI version")
	assert.Contains(t, stdoutBuff.String(), "Usage:", "porter command output should be printed")

	contents, err := os.ReadFile(filepath.Join(home, "config.toml"))
	require.NoError(t, err)
	wantTomlContents := "# I am a porter config"
	assert.Equal(t, wantTomlContents, string(contents))
	assert.Contains(t, gotStderr, wantTomlContents, "config file contents should be printed to stderr")

	contents, err = os.ReadFile(filepath.Join(home, "config.json"))
	require.NoError(t, err)
	wantJsonContents := "{}"
	assert.Equal(t, wantJsonContents, string(contents))
	assert.Contains(t, gotStderr, wantJsonContents, "config file contents should be printed to stderr")

	contents, err = os.ReadFile(filepath.Join(home, "a-binary"))
	require.NoError(t, err)
	wantBinaryContents := "binary contents"
	assert.Equal(t, wantBinaryContents, string(contents))
	assert.NotContains(t, gotStderr, wantBinaryContents, "binary file contents should NOT be printed")

	_, err = os.Stat(filepath.Join(home, ".hidden"))
	require.True(t, os.IsNotExist(err), "hidden files should not be copied")
}

func makeTestPorterHome(t *testing.T) string {
	home, err := ioutil.TempDir("", "porter-home")
	require.NoError(t, err)
	require.NoError(t, shx.Copy("../../bin/porter", home))
	return home
}
