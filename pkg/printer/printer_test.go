package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestPrintOptions_Validate(t *testing.T) {
	t.Run("allowed format", func(t *testing.T) {
		allowed := Formats{FormatPlaintext, FormatJson}
		opts := PrintOptions{RawFormat: "plaintext"}
		err := opts.Validate(allowed)
		require.NoError(t, err, "Validate should succeed for an allowed value")
		assert.Equal(t, FormatPlaintext, opts.Format, "Validate should set the Format field to the parsed format")
	})

	t.Run("unallowed format", func(t *testing.T) {
		allowed := Formats{FormatPlaintext, FormatJson}
		opts := PrintOptions{RawFormat: "yaml"}
		err := opts.Validate(allowed)
		require.EqualError(t, err, "invalid format: yaml", "Validate should fail for an unallowed value")
	})
}

func TestFormats_String(t *testing.T) {
	t.Run("zero length", func(t *testing.T) {
		f := Formats{}
		assert.Equal(t, "", f.String())
	})

	t.Run("single length", func(t *testing.T) {
		f := Formats{FormatPlaintext}
		assert.Equal(t, "plaintext", f.String())
	})

	t.Run("multi length", func(t *testing.T) {
		f := Formats{FormatPlaintext, FormatJson}
		assert.Equal(t, "plaintext, json", f.String())
	})
}
