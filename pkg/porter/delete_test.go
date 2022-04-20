package porter

import (
	"context"
	"testing"

	"get.porter.sh/porter/pkg/claims"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteInstallation(t *testing.T) {
	ctx := context.Background()

	testcases := []struct {
		name                string
		lastAction          string
		lastActionStatus    string
		force               bool
		installationRemains bool
		wantError           string
	}{
		{
			name:      "not yet installed",
			wantError: "not found",
		}, {
			name:                "last action not uninstall; no --force",
			lastAction:          "install",
			lastActionStatus:    cnab.StatusSucceeded,
			installationRemains: true,
			wantError:           ErrUnsafeInstallationDeleteRetryForce.Error(),
		}, {
			name:                "last action failed uninstall; no --force",
			lastAction:          "uninstall",
			lastActionStatus:    cnab.StatusFailed,
			installationRemains: true,
			wantError:           ErrUnsafeInstallationDeleteRetryForce.Error(),
		}, {
			name:             "last action not uninstall; --force",
			lastAction:       "install",
			lastActionStatus: cnab.StatusSucceeded,
			force:            true,
		}, {
			name:             "last action failed uninstall; --force",
			lastAction:       "uninstall",
			lastActionStatus: cnab.StatusFailed,
			force:            true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Teardown()

			var err error

			// Create test claim
			if tc.lastAction != "" {
				i := p.TestClaims.CreateInstallation(claims.NewInstallation("", "test"))
				c := p.TestClaims.CreateRun(i.NewRun(tc.lastAction))
				_ = p.TestClaims.CreateResult(c.NewResult(tc.lastActionStatus))
			}

			opts := DeleteOptions{}
			opts.Name = "test"
			opts.Force = tc.force

			err = p.DeleteInstallation(ctx, opts)
			if tc.wantError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantError)
			} else {
				require.NoError(t, err, "expected DeleteInstallation to succeed")
			}

			_, err = p.Claims.GetInstallation(ctx, "", "test")
			if tc.installationRemains {
				require.NoError(t, err, "expected installation to exist")
			} else {
				require.ErrorIs(t, err, storage.ErrNotFound{})
			}
		})
	}
}
