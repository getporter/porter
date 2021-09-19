package tests

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

// RequireErrorContains fails the test when the error doesn't contain
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