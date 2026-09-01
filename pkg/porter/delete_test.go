package porter

import (
	"context"
	"testing"

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
		referenced          bool
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
		}, {
			name:             "successful uninstall; no --force",
			lastAction:       "uninstall",
			lastActionStatus: cnab.StatusSucceeded,
		}, {
			name:                "still referenced; no --force",
			lastAction:          "uninstall",
			lastActionStatus:    cnab.StatusSucceeded,
			referenced:          true,
			installationRemains: true,
			wantError:           ErrInstallationReferencedRetryForce.Error(),
		}, {
			name:             "still referenced; --force",
			lastAction:       "uninstall",
			lastActionStatus: cnab.StatusSucceeded,
			referenced:       true,
			force:            true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewTestPorter(t)
			defer p.Close()

			var err error

			// Create test claim
			if tc.lastAction != "" {
				i := p.TestInstallations.CreateInstallation(storage.NewInstallation("", "test"), func(i *storage.Installation) {
					i.Status.Action = tc.lastAction
					i.Status.ResultStatus = tc.lastActionStatus
					if tc.referenced {
						i.AddReference("/parent", "db")
					}
				})
				c := p.TestInstallations.CreateRun(i.NewRun(tc.lastAction, cnab.ExtendedBundle{}))
				_ = p.TestInstallations.CreateResult(c.NewResult(tc.lastActionStatus))
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

			_, err = p.Installations.GetInstallation(ctx, "", "test")
			if tc.installationRemains {
				require.NoError(t, err, "expected installation to exist")
			} else {
				require.ErrorIs(t, err, storage.ErrNotFound{})
			}
		})
	}
}

// TestDeleteInstallation_SweepsReferences verifies that deleting an
// installation that was itself referencing other installations (to satisfy
// its own dependencies) removes its reference from those installations, so
// they don't consider themselves referenced by a record that no longer
// exists.
func TestDeleteInstallation_SweepsReferences(t *testing.T) {
	ctx := context.Background()

	p := NewTestPorter(t)
	defer p.Close()

	p.TestInstallations.CreateInstallation(storage.NewInstallation("", "mysqldb"), func(i *storage.Installation) {
		i.AddReference("/myinfra", "db")
	})

	parent := p.TestInstallations.CreateInstallation(storage.NewInstallation("", "myinfra"), func(i *storage.Installation) {
		i.Status.Action = cnab.ActionUninstall
		i.Status.ResultStatus = cnab.StatusSucceeded
	})
	c := p.TestInstallations.CreateRun(parent.NewRun(cnab.ActionUninstall, cnab.ExtendedBundle{}))
	_ = p.TestInstallations.CreateResult(c.NewResult(cnab.StatusSucceeded))

	opts := DeleteOptions{}
	opts.Name = "myinfra"

	err := p.DeleteInstallation(ctx, opts)
	require.NoError(t, err, "expected DeleteInstallation to succeed")

	updatedDep, err := p.Installations.GetInstallation(ctx, "", "mysqldb")
	require.NoError(t, err, "expected the referenced installation to still exist")
	assert.False(t, updatedDep.IsReferenced(), "expected the deleted installation's reference to be swept from mysqldb")
}
