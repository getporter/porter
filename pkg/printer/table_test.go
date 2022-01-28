package printer

import (
	"bytes"
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/require"
)

type testType struct {
	A, B interface{}
}

func printTestType(r interface{}) []string {
	row, ok := r.(testType)
	if !ok {
		panic("invalid row data passed to printTestType, should be testType")
	}
	return []string{fmt.Sprintf("%v", row.A), fmt.Sprintf("%v", row.B)}
}

func TestPrintTable(t *testing.T) {
	v := []testType{
		{A: "foo", B: "a really long bit of text that really should be wrapped nicely so that the entire table doesn't expand infinitely wide"},
		{A: "baz", B: "qux"},
		{A: 123, B: true},
	}

	b := &bytes.Buffer{}

	err := PrintTable(b, v, printTestType,
		"A", "B")

	require.NoError(t, err)
	test.CompareGoldenFile(t, "testdata/table-with-headers.txt", b.String())
}

func TestPrintTable_WithoutHeaders(t *testing.T) {
	v := []testType{
		{A: "foo", B: "bar"},
	}

	b := &bytes.Buffer{}

	err := PrintTable(b, v, printTestType)

	require.NoError(t, err)
	test.CompareGoldenFile(t, "testdata/table-without-headers.txt", b.String())
}
