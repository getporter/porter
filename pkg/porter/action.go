package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/claims"
)

// ExecuteAction runs the specified action. Supported actions are: install, upgrade, invoke.
// The uninstall action works in reverse so it's implemented separately.
func (p *Porter) ExecuteAction(installation claims.Installation, action BundleAction) error {
	actionOpts := action.GetOptions()

	actionArgs, err := p.BuildActionArgs(installation, action)
	if err != nil {
		return err
	}

	deperator := newDependencyExecutioner(p, action, actionArgs)
	err = deperator.Prepare()
	if err != nil {
		return err
	}

	err = deperator.Execute()
	if err != nil {
		return err
	}

	deperator.PrepareRootActionArguments(&actionArgs)

	fmt.Fprintf(p.Out, "%s %s...\n", action.GetActionVerb(), actionOpts.Name)
	return p.CNAB.Execute(actionArgs)
}
