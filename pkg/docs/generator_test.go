package docs

import (
	"fmt"
	"testing"

	"get.porter.sh/porter/pkg/porter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateDocsCommand(t *testing.T) {

	testcases := []struct {
		name        string
		destination string
		wantError   string
	}{
		{"should return error if destination doesn't exist", "./no-existing/destination/directory", "--destination %q doesn't exist"},
		{"should not return error if destination exists", ".", ""},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := porter.NewTestPorter(t)
			opts := DocsOptions{
				Destination: tc.destination,
			}
			err := opts.Validate(p.Context)
			if tc.wantError == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, err.Error(), fmt.Sprintf(tc.wantError, tc.destination))
			}
		})
	}

}
