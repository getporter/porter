package porter

import (
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionOptions_Validate(t *testing.T) {
	testcases := []struct {
		name       string
		opts       VersionOptions
		wantFormat printer.Format
		wantError  string
	}{
		{"json", VersionOptions{printer.PrintOptions{RawFormat: "json"}}, printer.FormatJson, ""},
		{"plaintext", VersionOptions{printer.PrintOptions{RawFormat: "plaintext"}}, printer.FormatPlaintext, ""},
		{"default", VersionOptions{}, printer.FormatPlaintext, ""},
		{"yaml - unsupported", VersionOptions{printer.PrintOptions{RawFormat: "yaml"}}, printer.Format(""), "unsupported"},
		{"oops - invalid", VersionOptions{printer.PrintOptions{RawFormat: "oops"}}, printer.Format(""), "invalid"},
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

func TestPrintVersion(t *testing.T) {
	pkg.Commit = "abc123"
	pkg.Version = "v1.2.3"

	p := NewTestPorter(t)

	opts := VersionOptions{}
	err := opts.Validate()
	require.NoError(t, err)
	p.PrintVersion(opts)

	gotOutput := p.TestConfig.TestContext.GetOutput()
	wantOutput := "porter v1.2.3 (abc123)"
	if !strings.Contains(gotOutput, wantOutput) {
		t.Fatalf("invalid output:\nWANT:\t%q\nGOT:\t%q\n", wantOutput, gotOutput)
	}
}
