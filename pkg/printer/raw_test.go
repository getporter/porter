package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintRaw(t *testing.T) {
	v := struct {
		A bool
		B string
	}{
		A: false,
		B: "waltz",
	}

	b := &bytes.Buffer{}
	err := PrintRaw(b, v)

	require.Nil(t, err)
	require.Equal(t, "{false waltz}\n", b.String())
}
