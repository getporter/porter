package context

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext_EnvironMap(t *testing.T) {
	c := NewTestContext(t)
	c.Clearenv()

	c.Setenv("a", "1")
	c.Setenv("b", "2")

	got := c.EnvironMap()

	want := map[string]string{
		"a": "1",
		"b": "2",
	}
	assert.Equal(t, want, got)

	// Make sure we have a copy
	got["c"] = "3"
	assert.Empty(t, c.Getenv("c"), "Expected to get a copy of the context's environment variables")
}

func TestContext_SetSensitiveValues(t *testing.T) {
	c := NewTestContext(t)
	c.ShowSensitiveValues = true

	td := []string{
		"key=sensitive_value",
		"another=sensitive_value",
	}

	c.SetSensitiveValues(td)
	want := strings.Join(td, "\n") + "\n"
	got := c.capturedOut.String()
	assert.Equal(t, want, got)

}
