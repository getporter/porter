package version

import (
	"testing"

	"github.com/deislabs/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionOptions_Validate(t *testing.T) {
	testcases := []struct {
		name       string
		opts       Options
		wantFormat printer.Format
		wantError  string
	}{
		{"json", Options{printer.PrintOptions{RawFormat: "json"}}, printer.FormatJson, ""},
		{"plaintext", Options{printer.PrintOptions{RawFormat: "plaintext"}}, printer.FormatPlaintext, ""},
		{"default", Options{}, printer.FormatPlaintext, ""},
		{"yaml - unsupported", Options{printer.PrintOptions{RawFormat: "yaml"}}, printer.Format(""), "unsupported"},
		{"oops - invalid", Options{printer.PrintOptions{RawFormat: "oops"}}, printer.Format(""), "invalid"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.opts.Validate()

			if tc.wantError == "" {
				require.NoError(t, err, "Validate should not have returned an error")
			} else {
				require.Error(t, err, "Validate should have returned an error")
				assert.Contains(t, err.Error(), tc.wantError, "Validate did not return the expected error message")
			}
		})
	}
}
