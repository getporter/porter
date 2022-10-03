package docs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/carolynvs/magex/shx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureOperatorRepository(t *testing.T) {
	t.Run("has local repo", func(t *testing.T) {
		tmp, err := ioutil.TempDir("", "porter-docs-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmp)

		resolvedPath, err := ensureOperatorRepositoryIn(tmp, "")
		require.NoError(t, err)
		require.Equal(t, tmp, resolvedPath)
	})

	t.Run("missing local repo", func(t *testing.T) {
		tmp, err := ioutil.TempDir("", "porter-docs-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmp)

		resolvedPath, err := ensureOperatorRepositoryIn("missing", tmp)
		require.NoError(t, err)
		require.Equal(t, tmp, resolvedPath)
	})

	t.Run("local repo unset", func(t *testing.T) {
		tmp, err := ioutil.TempDir("", "porter-docs-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmp)

		resolvedPath, err := ensureOperatorRepositoryIn("", tmp)
		require.NoError(t, err)
		require.Equal(t, tmp, resolvedPath)
	})

	t.Run("empty default path clones repo", func(t *testing.T) {
		tmp, err := ioutil.TempDir("", "porter-docs-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmp)

		resolvedPath, err := ensureOperatorRepositoryIn("", tmp)
		require.NoError(t, err)
		require.Equal(t, tmp, resolvedPath)

		err = shx.Command("git", "status").In(resolvedPath).RunE()
		require.NoError(t, err, "clone failed")
	})

	t.Run("changes in default path are reset", func(t *testing.T) {
		tmp, err := ioutil.TempDir("", "porter-docs-test")
		require.NoError(t, err)
		defer os.RemoveAll(tmp)

		repoPath, err := ensureOperatorRepositoryIn("", tmp)
		require.NoError(t, err)

		// make a change
		readme := filepath.Join(repoPath, "README.md")
		require.NoError(t, os.Remove(readme))

		// Make sure rerunning resets the change
		_, err = ensureOperatorRepositoryIn("", tmp)
		require.NoError(t, err)
		require.FileExists(t, readme)
	})
}

func Test_setPullRequestBaseURL(t *testing.T) {
	os.Setenv("DEPLOY_PRIME_URL", "https://preview--porter.netlify.app")
	setPullRequestBaseURL()
	assert.Equal(t, "https://preview--porter.netlify.app/", os.Getenv("BASEURL"))
}

func TestDocsBranchPreview(t *testing.T) {
	os.Setenv("BRANCH", "release/v1")
	setBranchBaseURL()
	assert.Equal(t, "https://release-v1.getporter.org/", os.Getenv("BASEURL"))
}
