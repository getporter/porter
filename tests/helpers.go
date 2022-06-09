package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertFilePermissionsEqual checks that the file permission bits are equal for the specified file.
func AssertFilePermissionsEqual(t *testing.T, path string, want os.FileMode, got os.FileMode) bool {
	want = want.Perm()
	got = got.Perm()
	return assert.Equal(t, want, got,
		"incorrect file permissions on %s. want: %o, got %o", path, want, got)
}

// AssertDirectoryPermissionsEqual checks that all files in the specified path match the desired
// file mode. Uses a glob pattern to match.
func AssertDirectoryPermissionsEqual(t *testing.T, path string, mode os.FileMode) bool {
	files, err := filepath.Glob(path)
	if os.IsNotExist(err) {
		return true
	}
	require.NoError(t, err)

	for _, file := range files {
		info, err := os.Stat(file)
		require.NoError(t, err)
		if info.IsDir() {
			continue
		}
		if !AssertFilePermissionsEqual(t, path, mode, info.Mode()) {
			return false
		}
	}

	return true
}
