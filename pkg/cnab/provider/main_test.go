package cnabprovider

import (
	"testing"

	"get.porter.sh/porter/pkg/test"
)

func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}
