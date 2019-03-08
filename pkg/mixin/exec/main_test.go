package exec

import (
	"testing"

	"github.com/deislabs/porter/pkg/test"
)

// sad hack: not sure how to make a common test main for all my subpackages
func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}
