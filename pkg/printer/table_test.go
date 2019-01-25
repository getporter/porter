package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

type testType struct {
	A, B string
}

func TestPrintTable(t *testing.T) {
	v := []testType{
		{A: "foo", B: "bar"},
		{A: "baz", B: "qux"},
	}

	b := &bytes.Buffer{}

	err := PrintTable(b, v, func(r interface{}) []interface{} {
		row, ok := r.(testType)
		require.True(t, ok)
		return []interface{}{row.A, row.B}
	},
		"A", "B")

	require.Nil(t, err)
	require.Equal(t, "A B\nfoo bar\nbaz qux\n", b.String())
}

func TestPrintTable_WithoutHeaders(t *testing.T) {
	v := []testType{
		{A: "foo", B: "bar"},
	}

	b := &bytes.Buffer{}

	err := PrintTable(b, v, func(r interface{}) []interface{} {
		row, ok := r.(testType)
		require.True(t, ok)
		return []interface{}{row.A, row.B}
	})

	require.Nil(t, err)
	require.Equal(t, "foo bar\n", b.String())
}
