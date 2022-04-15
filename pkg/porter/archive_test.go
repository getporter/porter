package porter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchive_ParentDirDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	opts := ArchiveOptions{}
	opts.Reference = "myreg/mybuns:v0.1.0"

	err := opts.Validate(context.Background(), []string{"/path/to/file"}, p.Porter)
	require.NoError(t, err, "expected no validation error to occur")

	err = p.Archive(context.Background(), opts)
	require.EqualError(t, err, "parent directory \"/path/to\" does not exist")
}

func TestArchive_Validate(t *testing.T) {
	p := NewTestPorter(t)
	defer p.Teardown()

	testcases := []struct {
		name      string
		args      []string
		reference string
		wantError string
	}{
		{"no arg", nil, "", "destination file is required"},
		{"no tag", []string{"/path/to/file"}, "", "must provide a value for --reference of the form REGISTRY/bundle:tag"},
		{"too many args", []string{"/path/to/file", "moar args!"}, "myreg/mybuns:v0.1.0", "only one positional argument may be specified, the archive file name, but multiple were received: [/path/to/file moar args!]"},
		{"just right", []string{"/path/to/file"}, "myreg/mybuns:v0.1.0", ""},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := ArchiveOptions{}
			opts.Reference = tc.reference

			err := opts.Validate(context.Background(), tc.args, p.Porter)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err, "expected no validation error to occur")
			}
		})
	}
}
