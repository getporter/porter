package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

type Test struct {
	value          interface{}
	expectedOutput string
}

type specialType struct {
	A bool
	B float32
}

func TestPrintPlaintext(t *testing.T) {
	tests := []Test{
		Test{
			value:          "I'm a string",
			expectedOutput: "I'm a string",
		},
		Test{
			value:          []int{1, 2, 3},
			expectedOutput: "[1 2 3]",
		},
		Test{
			value:          specialType{A: true, B: 123},
			expectedOutput: "{true 123}",
		},
	}

	for _, test := range tests {
		b := &bytes.Buffer{}
		err := PrintPlaintext(b, test.value)

		require.Nil(t, err)
		require.Equal(t, test.expectedOutput, b.String())
	}
}
