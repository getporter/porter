package porter

import (
	"fmt"

	"github.com/pkg/errors"
)

// ExecuteAction runs the specified action. Supported actions are: install, upgrade, invoke.
// The uninstall action works in reverse so it's implemented separately.
func (p *Porter) ExecuteAction(action BundleAction) error {
	actionOpts := action.GetOptions()

	err := p.prepullBundleByReference(actionOpts)
	if err != nil {
		return errors.Wrap(err, "unable to pull bundle before installation")
	}

	err = p.ensureLocalBundleIsUpToDate(actionOpts.bundleFileOptions)
	if err != nil {
		return err
	}

	deperator := newDependencyExecutioner(p, action.GetAction())
	err = deperator.Prepare(action)
	if err != nil {
		return err
	}

	err = deperator.Execute()
	if err != nil {
		return err
	}

	actionArgs, err := p.BuildActionArgs(action)
	if err != nil {
		return err
	}
	deperator.PrepareRootActionArguments(&actionArgs)

	fmt.Fprintf(p.Out, "%s %s...\n", action.GetActionVerb(), actionOpts.Name)
	return p.CNAB.Execute(actionArgs)
}
