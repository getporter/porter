package tests

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RequireErrorContains fails the test when the error doesn't contain
// AssertFilePermissionsEqual checks that the file permission bits are equal for the specified file.
func AssertFilePermissionsEqual(t *testing.T, path string, want os.FileMode, got os.FileMode) bool {
	want = want.Perm()
	got = got.Perm()
	return assert.Equal(t, want, got|want,
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

// the specified substring. This is less fragile than using require.EqualError
func RequireErrorContains(t *testing.T, err error, substring string, msgAndArgs ...interface{}) {
	require.Error(t, err)
	require.Contains(t, err.Error(), substring, msgAndArgs...)
}

// GenerateDatabaseName comes up with a valid mongodb database name from a Go test name.
func GenerateDatabaseName(testName string) string {
	reg, err := regexp.Compile(`[^a-zA-Z0-9_]+`)
	if err != nil {
		panic(err)
	}
	safeTestName := reg.ReplaceAllString(testName, "_")
	if len(safeTestName) > 50 {
		safeTestName = fmt.Sprintf("%x", md5.Sum([]byte(safeTestName)))
	}
	return fmt.Sprintf("porter_%s", safeTestName)
}

// This is the same as require.Contains but it prints the output string without
// newlines escaped so that it's easier to read.
func RequireOutputContains(t *testing.T, output string, substring string, msgAndArgs ...interface{}) {
	ok := assert.Contains(t, output, substring, msgAndArgs...)
	if !ok {
		t.Errorf("%s\ndoes not contain\n%s", output, substring)
		t.FailNow()
	}
}
