package exec

import (
	"testing"

	"get.porter.sh/porter/pkg/test"
)

// sad hack: not sure how to make a common test main for all my subpackages
func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}
