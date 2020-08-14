package kubernetes

import (
	"testing"

	"get.porter.sh/porter/pkg/context"
)

type TestMixin struct {
	*Mixin
	TestContext *context.TestContext
}

func NewTestMixin(t *testing.T) *TestMixin {
	c := context.NewTestContext(t)
	m := New()
	m.Context = c.Context
	return &TestMixin{
		Mixin:       m,
		TestContext: c,
	}
}

// trimQuotes receives data and returns it with double or single wrapping quotes removed
func trimQuotes(data []byte) []byte {
	if len(data) >= 2 {
		if c := data[len(data)-1]; data[0] == c && (c == '"' || c == '\'') {
			return data[1 : len(data)-1]
		}
	}
	return data
}
