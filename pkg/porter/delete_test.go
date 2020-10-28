package porter

import (
	"testing"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/claim"
	"github.com/stretchr/testify/require"
)

func TestDeleteInstallation(t *testing.T) {
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
			wantError: "unable to read status for installation test: Installation does not exist",
		}, {
			name:                "last action not uninstall; no --force",
			lastAction:          "install",
			lastActionStatus:    claim.StatusSucceeded,
			installationRemains: true,
			wantError:           ErrUnsafeInstallationDeleteRetryForce.Error(),
		}, {
			name:                "last action failed uninstall; no --force",
			lastAction:          "uninstall",
			lastActionStatus:    claim.StatusFailed,
			installationRemains: true,
			wantError:           ErrUnsafeInstallationDeleteRetryForce.Error(),
		}, {
			name:             "last action not uninstall; --force",
			lastAction:       "install",
			lastActionStatus: claim.StatusSucceeded,
			force:            true,
		}, {
			name:             "last action failed uninstall; --force",
			lastAction:       "uninstall",
			lastActionStatus: claim.StatusFailed,
			force:            true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			p.TestConfig.SetupPorterHome()

			var err error

			// Create test claim
			if tc.lastAction != "" {
				c := p.TestClaims.CreateClaim("test", tc.lastAction, bundle.Bundle{}, nil)
				_ = p.TestClaims.CreateResult(c, tc.lastActionStatus)
			}

			opts := DeleteOptions{}
			opts.Name = "test"
			opts.Force = tc.force

			err = p.DeleteInstallation(opts)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err, "expected DeleteInstallation to succeed")
			}

			_, err = p.Claims.ReadInstallation("test")
			if tc.installationRemains {
				require.NoError(t, err, "expected installation to exist")
			} else {
				require.EqualError(t, err, "Installation does not exist")
			}
		})
	}
}
