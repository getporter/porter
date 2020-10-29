package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintYaml(t *testing.T) {
	v := struct {
		A string
		B bool
	}{
		A: "foo",
		B: true,
	}

	b := &bytes.Buffer{}
	err := PrintYaml(b, v)

	require.Nil(t, err)
	require.Equal(t, "a: foo\nb: true\n\n", b.String())
}
