package mage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPickBranchName(t *testing.T) {
	// These aren't set locally but are set on a CI run
	os.Unsetenv("SYSTEM_PULLREQUEST_SOURCEBRANCH")
	os.Unsetenv("BUILD_SOURCEBRANCHNAME")

	t.Run("origin/main", func(t *testing.T) {
		refs := []string{
			"refs/heads/foo",
			"refs/remotes/origin/main",
			"refs/remotes/origin/8252b6e4b1983702c7387ece7f971ef74047b746",
			"refs/tags/v0.38.3",
		}
		branch := pickBranchName(refs)
		assert.Equal(t, "main", branch)
	})

	t.Run("main", func(t *testing.T) {
		refs := []string{
			"refs/heads/foo",
			"refs/heads/main",
			"refs/remotes/origin/8252b6e4b1983702c7387ece7f971ef74047b746",
			"refs/tags/v0.38.3",
		}
		branch := pickBranchName(refs)
		assert.Equal(t, "main", branch)
	})

	t.Run("pull request", func(t *testing.T) {
		os.Setenv("SYSTEM_PULLREQUEST_SOURCEBRANCH", "patch-1")
		defer os.Unsetenv("SYSTEM_PULLREQUEST_SOURCEBRANCH")

		refs := []string{
			"refs/remotes/origin/foo",
			"refs/remotes/origin/8252b6e4b1983702c7387ece7f971ef74047b746",
		}
		branch := pickBranchName(refs)
		assert.Equal(t, "dev", branch)
	})

	t.Run("branch build", func(t *testing.T) {
		os.Setenv("BUILD_SOURCEBRANCHNAME", "foo")
		defer os.Unsetenv("BUILD_SOURCEBRANCHNAME")

		refs := []string{
			"refs/remotes/origin/foo",
			"refs/remotes/origin/8252b6e4b1983702c7387ece7f971ef74047b746",
		}
		branch := pickBranchName(refs)
		assert.Equal(t, "dev", branch)
	})

	t.Run("tagged release on v1", func(t *testing.T) {
		refs := []string{
			"refs/heads/release/v1",
			"refs/remotes/origin/8252b6e4b1983702c7387ece7f971ef74047b746",
			"refs/tags/v1.0.0-alpha.1",
		}
		branch := pickBranchName(refs)
		assert.Equal(t, "v1", branch)
	})

	t.Run("tagged release on main", func(t *testing.T) {
		refs := []string{
			"refs/heads/release/v1",
			"refs/heads/main",
			"refs/remotes/origin/8252b6e4b1983702c7387ece7f971ef74047b746",
			"refs/tags/v0.38.3",
		}
		branch := pickBranchName(refs)
		assert.Equal(t, "main", branch)
	})

	t.Run("local branch", func(t *testing.T) {
		refs := []string{
			"refs/heads/foo",
		}
		branch := pickBranchName(refs)
		assert.Equal(t, "dev", branch)
	})
}
