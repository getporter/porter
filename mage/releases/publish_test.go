package releases

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestAppendDataPath(t *testing.T) {
	data := make([]byte, 10)
	_, err := rand.Read(data)
	require.NoError(t, err)
	dataPath := "test/random"
	expected := hex.EncodeToString(data) + "  random"

	output := AppendDataPath(data, dataPath)
	require.Equal(t, expected, output)
}
