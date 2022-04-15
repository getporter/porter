package porter

import (
	"context"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

var _ BundleAction = NewUpgradeOptions()

// UpgradeOptions that may be specified when upgrading a bundle.
// Porter handles defaulting any missing values.
type UpgradeOptions struct {
	*BundleActionOptions

	// Version of the bundle to upgrade to
	Version string
}

func NewUpgradeOptions() UpgradeOptions {
	return UpgradeOptions{BundleActionOptions: &BundleActionOptions{}}
}

func (o UpgradeOptions) Validate(ctx context.Context, args []string, p *Porter) error {
	if o.Version != "" && o.Reference != "" {
		return errors.New("either --version or --reference may be set, but not both")
	}

	if o.Version != "" {
		v, err := semver.NewVersion(o.Version)
		if err != nil {
			return errors.New("invalid bundle version --version. Must be a semantic version, for example 1.2.3")
		}
		o.Version = v.String()
	}

	return o.BundleActionOptions.Validate(ctx, args, p)
}

func (o UpgradeOptions) GetAction() string {
	return cnab.ActionUpgrade
}

func (o UpgradeOptions) GetActionVerb() string {
	return "upgrading"
}

// UpgradeBundle accepts a set of pre-validated UpgradeOptions and uses
// them to upgrade a bundle.
func (p *Porter) UpgradeBundle(ctx context.Context, opts UpgradeOptions) error {
	// Figure out which bundle/installation we are working with
	_, err := p.resolveBundleReference(ctx, opts.BundleActionOptions)
	if err != nil {
		return err
	}

	// Sync any changes specified by the user to the installation before running upgrade
	i, err := p.Claims.GetInstallation(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return errors.Wrapf(err, "could not find installation %s/%s", opts.Namespace, opts.Name)
	}

	if opts.Reference != "" {
		i.TrackBundle(opts.GetReference())
	} else if opts.Version != "" {
		i.Bundle.Version = opts.Version
		i.Bundle.Digest = ""
		i.Bundle.Tag = ""
	}

	err = p.applyActionOptionsToInstallation(ctx, &i, opts.BundleActionOptions)
	i.Status.Modified = time.Now()
	err = i.Validate()
	if err != nil {
		return err
	}
	err = p.Claims.UpdateInstallation(ctx, i)
	if err != nil {
		return err
	}

	// Re-resolve the bundle after we have figured out the version we are upgrading to
	opts.bundleRef = nil
	_, err = p.resolveBundleReference(ctx, opts.BundleActionOptions)
	if err != nil {
		return err
	}

	return p.ExecuteAction(ctx, i, opts)
}
