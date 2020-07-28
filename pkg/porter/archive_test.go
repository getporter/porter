package porter

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchive_ParentDirDoesNotExist(t *testing.T) {
	p := NewTestPorter(t)

	opts := ArchiveOptions{}

	err := opts.Validate([]string{"/path/to/file"}, p.Porter)
	require.NoError(t, err, "expected no validation error to occur")

	err = p.Archive(opts)
	require.EqualError(t, err, "parent directory \"/path/to\" does not exist")
}
