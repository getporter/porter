package porter

import (
	"context"
	"errors"
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/Masterminds/semver/v3"
)

var _ BundleAction = NewUpgradeOptions()

// UpgradeOptions that may be specified when upgrading a bundle.
// Porter handles defaulting any missing values.
type UpgradeOptions struct {
	*BundleExecutionOptions

	// Version of the bundle to upgrade to
	Version string
}

func NewUpgradeOptions() *UpgradeOptions {
	return &UpgradeOptions{
		BundleExecutionOptions: NewBundleExecutionOptions(),
	}
}

func (o *UpgradeOptions) Validate(ctx context.Context, args []string, p *Porter) error {
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

	return o.BundleExecutionOptions.Validate(ctx, args, p)
}

func (o *UpgradeOptions) GetAction() string {
	return cnab.ActionUpgrade
}

func (o *UpgradeOptions) GetActionVerb() string {
	return "upgrading"
}

// UpgradeBundle accepts a set of pre-validated UpgradeOptions and uses
// them to upgrade a bundle.
func (p *Porter) UpgradeBundle(ctx context.Context, opts *UpgradeOptions) error {
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	// Sync any changes specified by the user to the installation before running upgrade
	i, err := p.Installations.GetInstallation(ctx, opts.Namespace, opts.Name)
	if err != nil {
		return span.Errorf("could not find installation %s/%s: %w", opts.Namespace, opts.Name, err)
	}

	if opts.Reference != "" {
		i.TrackBundle(opts.GetReference())
	} else if opts.Version != "" {
		i.Bundle.Version = opts.Version
		i.Bundle.Digest = ""
		i.Bundle.Tag = ""
	}

	err = p.applyActionOptionsToInstallation(ctx, opts, &i)
	if err != nil {
		return span.Errorf("could not apply options to installation: %w", err)
	}
	i.Status.Modified = time.Now()

	checkStrategy := p.GetSchemaCheckStrategy(ctx)
	err = i.Validate(ctx, checkStrategy)
	if err != nil {
		return err
	}

	err = p.Installations.UpdateInstallation(ctx, i)
	if err != nil {
		return err
	}

	// Re-resolve the bundle after we have figured out the version we are upgrading to
	opts.UnsetBundleReference()
	if _, err := opts.GetBundleReference(ctx, p); err != nil {
		return err
	}

	return p.ExecuteAction(ctx, i, opts)
}
