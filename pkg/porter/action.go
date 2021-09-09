package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/claims"
)

// ExecuteAction runs the specified action. Supported actions are: install, upgrade, invoke.
// The uninstall action works in reverse so it's implemented separately.
func (p *Porter) ExecuteAction(installation claims.Installation, action BundleAction) error {
	actionOpts := action.GetOptions()

	deperator := newDependencyExecutioner(p, installation, action)
	err := deperator.Prepare()
	if err != nil {
		return err
	}

	err = deperator.Execute()
	if err != nil {
		return err
	}

	actionArgs, err := deperator.PrepareRootActionArguments()
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "%s %s...\n", action.GetActionVerb(), actionOpts.Name)
	return p.CNAB.Execute(actionArgs)
}
