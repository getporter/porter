package printer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFormat(t *testing.T) {
	testcases := map[string]bool{
		"table": true,
		"json":  true,
		"human": false,
	}

	for name, valid := range testcases {
		t.Run(name, func(t *testing.T) {
			opts := PrintOptions{
				RawFormat: name,
			}

			err := opts.ParseFormat()
			if valid {
				require.Nil(t, err)
				require.Equal(t, name, string(opts.Format))
			} else {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), "invalid format")
			}
		})
	}
}
