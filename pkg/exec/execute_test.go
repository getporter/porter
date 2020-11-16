package exec

import (
	"bytes"
	"io/ioutil"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_Execute(t *testing.T) {
	m := NewTestMixin(t)

	err := m.FileSystem.WriteFile("config.txt", []byte("abc123"), 0644)
	require.NoError(t, err)

	stdin, err := ioutil.ReadFile("testdata/outputs.yaml")
	require.NoError(t, err)
	m.In = bytes.NewBuffer(stdin)

	m.Setenv(test.ExpectedCommandEnv, "foo")

	err = m.Execute(ExecuteOptions{})
	require.NoError(t, err, "Execute should not have returned an error")

	exists, _ := m.FileSystem.Exists("/cnab/app/porter/outputs/file")
	assert.True(t, exists, "file output was not evaluated")

	exists, _ = m.FileSystem.Exists("/cnab/app/porter/outputs/regex")
	assert.True(t, exists, "regex output was not evaluated")

	exists, _ = m.FileSystem.Exists("/cnab/app/porter/outputs/jsonpath")
	assert.True(t, exists, "jsonpath output was not evaluated")
}
