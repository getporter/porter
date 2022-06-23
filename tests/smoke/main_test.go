//go:build smoke

package smoke

import (
	"testing"

	"get.porter.sh/porter/tests/tester"
	"github.com/stretchr/testify/require"
)

// Make sure the porter binary that we are using is okay
func TestPorterBinary(t *testing.T) {
	test, err := tester.NewTest(t)
	defer test.Close()
	require.NoError(t, err)

	test.RequirePorter("help")
	test.RequirePorter("version")
}
