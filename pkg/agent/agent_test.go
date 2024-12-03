//go:build integration

package agent

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uwu-tools/magex/shx"
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

	_, err = os.ReadFile(filepath.Join(home, "config.toml"))
	require.NoError(t, err)

	_, err = os.ReadFile(filepath.Join(home, "config.json"))
	require.NoError(t, err)

	_, err = os.ReadFile(filepath.Join(home, "a-binary"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(home, ".hidden"))
	require.True(t, os.IsNotExist(err), "hidden files should not be copied")
}

func makeTestPorterHome(t *testing.T) string {
	home, err := os.MkdirTemp("", "porter-home")
	require.NoError(t, err)
	porter_binary := "../../bin/porter"
	if runtime.GOOS == "windows" {
		porter_binary += ".exe"
	}
	require.NoError(t, shx.Copy(porter_binary, home))
	return home
}
