package printer

import (
	"testing"

	"get.porter.sh/porter/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormat(t *testing.T) {
	testcases := []struct {
		rawFormat  string
		wantFormat Format
		wantErr    string
	}{
		{rawFormat: "plaintext", wantFormat: FormatPlaintext},
		{rawFormat: "json", wantFormat: FormatJson},
		{rawFormat: "yaml", wantFormat: FormatYaml},
		{rawFormat: "", wantFormat: FormatPlaintext},
		{rawFormat: "human", wantErr: "invalid format"},
	}

	for _, tc := range testcases {
		t.Run(tc.rawFormat, func(t *testing.T) {
			opts := PrintOptions{
				RawFormat: tc.rawFormat,
			}

			err := opts.ParseFormat()
			if tc.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tc.wantFormat, opts.Format, "incorrect format was returned by ParseFormat")
			} else {
				tests.RequireErrorContains(t, err, tc.wantErr, "unexpected error returned by ParseFormat")
			}
		})
	}
}

func TestPrintOptions_Validate(t *testing.T) {
	t.Run("default format", func(t *testing.T) {
		allowed := Formats{FormatPlaintext, FormatJson}
		opts := PrintOptions{}
		err := opts.Validate(FormatJson, allowed)
		require.NoError(t, err, "Validate should succeed for a defaulted value")
		assert.Equal(t, FormatJson, opts.Format, "Validate should set the Format field to the default format")
	})

	t.Run("allowed format", func(t *testing.T) {
		allowed := Formats{FormatPlaintext, FormatJson}
		opts := PrintOptions{RawFormat: "plaintext"}
		err := opts.Validate("", allowed)
		require.NoError(t, err, "Validate should succeed for an allowed value")
		assert.Equal(t, FormatPlaintext, opts.Format, "Validate should set the Format field to the parsed format")
	})

	t.Run("unallowed format", func(t *testing.T) {
		allowed := Formats{FormatPlaintext, FormatJson}
		opts := PrintOptions{RawFormat: "yaml"}
		err := opts.Validate("", allowed)
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
