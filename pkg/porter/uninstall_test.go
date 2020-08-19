package porter

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestPorter_HandleUninstallErrs(t *testing.T) {
	testcases := []struct {
		name    string
		opts    UninstallDeleteOptions
		err     error
		wantOut string
		wantErr string
	}{
		{
			name: "no delete; no err",
			opts: UninstallDeleteOptions{},
		}, {
			name:    "no delete",
			opts:    UninstallDeleteOptions{},
			err:     errors.New("an error was encountered"),
			wantErr: "an error was encountered",
		}, {
			name:    "--delete; no --force-delete",
			opts:    UninstallDeleteOptions{Delete: true},
			err:     errors.New("an error was encountered"),
			wantErr: fmt.Sprintf("2 errors occurred:\n\t* an error was encountered\n\t* %s\n\n", ErrUnsafeInstallationDeleteRetryForceDelete),
		}, {
			name:    "--force-delete",
			opts:    UninstallDeleteOptions{ForceDelete: true},
			err:     errors.New("an error was encountered"),
			wantOut: "ignoring the following errors as --force-delete is true:\n  an error was encountered",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			out := bytes.NewBufferString("")

			err := tc.opts.handleUninstallErrs(out, tc.err)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantOut, out.String())
		})
	}
}
