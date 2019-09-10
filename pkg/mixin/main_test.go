package mixin

import (
	"testing"

	"github.com/deislabs/porter/pkg/test"
)

func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}
