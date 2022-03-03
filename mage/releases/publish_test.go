package releases

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddChecksumExt(t *testing.T) {
	tests := []struct {
		input         string
		expectedAdded bool
		expected      string
	}{
		{
			input:         "porter.sh",
			expectedAdded: true,
			expected:      "porter.sh.sha256sum",
		},
		{
			input:         "porter.sh.sha256sum",
			expectedAdded: false,
			expected:      "porter.sh.sha256sum",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			output, added := AddChecksumExt(tt.input)
			assert.Equal(t, tt.expected, output)
			assert.Equal(t, tt.expectedAdded, added)
		})
	}

}
