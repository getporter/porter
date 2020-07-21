package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/manifest"

	"github.com/pkg/errors"
)

// InstallOptions that may be specified when installing a bundle.
// Porter handles defaulting any missing values.
type InstallOptions struct {
	BundleLifecycleOpts
}

// InstallBundle accepts a set of pre-validated InstallOptions and uses
// them to install a bundle.
func (p *Porter) InstallBundle(opts InstallOptions) error {
	err := p.prepullBundleByTag(&opts.BundleLifecycleOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before installation")
	}

	err = p.ensureLocalBundleIsUpToDate(opts.bundleFileOptions)
	if err != nil {
		return err
	}

	deperator := newDependencyExecutioner(p)
	err = deperator.Prepare(opts.BundleLifecycleOpts, p.CNAB.Install)
	if err != nil {
		return err
	}

	err = deperator.Execute(manifest.ActionInstall)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "installing %s...\n", opts.Name)
	return p.CNAB.Install(opts.ToActionArgs(deperator))
}
