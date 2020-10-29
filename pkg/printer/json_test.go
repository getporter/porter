package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintJson(t *testing.T) {
	v := struct {
		A string
	}{
		A: "foo",
	}

	b := &bytes.Buffer{}
	err := PrintJson(b, v)

	require.Nil(t, err)
	// Make sure that it is printing pretty with proper indents and spaces and trailing newlines
	require.Equal(t, "{\n  \"A\": \"foo\"\n}\n", b.String())
}
