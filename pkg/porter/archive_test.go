package porter

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchive_ParentDirDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)

	opts := ArchiveOptions{}
	opts.Tag = "myreg/mybuns:v0.1.0"

	missingParent := filepath.Join(p.Getwd(), "missing")
	err := opts.Validate([]string{filepath.Join(missingParent, "somepath")}, p.Porter)
	require.NoError(t, err, "expected no validation error to occur")

	err = p.Archive(opts)
	require.Error(t, err, "expected Archive to fail")
	assert.Contains(t, err.Error(), fmt.Sprintf("parent directory %q does not exist", missingParent))
}

func TestArchive_Validate(t *testing.T) {
	p := NewTestPorter(t)

	testcases := []struct {
		name      string
		args      []string
		tag       string
		wantError string
	}{
		{"no arg", nil, "", "destination file is required"},
		{"no tag", []string{"/path/to/file"}, "", "must provide a value for --tag of the form REGISTRY/bundle:tag"},
		{"too many args", []string{"/path/to/file", "moar args!"}, "myreg/mybuns:v0.1.0", "only one positional argument may be specified, the archive file name, but multiple were received: [/path/to/file moar args!]"},
		{"just right", []string{"/path/to/file"}, "myreg/mybuns:v0.1.0", ""},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := ArchiveOptions{}
			opts.Tag = tc.tag

			err := opts.Validate(tc.args, p.Porter)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err, "expected no validation error to occur")
			}
		})
	}
}
