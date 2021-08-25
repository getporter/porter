package porter

import (
	"fmt"

	"get.porter.sh/porter/pkg/claims"
)

// ExecuteAction runs the specified action. Supported actions are: install, upgrade, invoke.
// The uninstall action works in reverse so it's implemented separately.
func (p *Porter) ExecuteAction(installation claims.Installation, action BundleAction) error {
	actionOpts := action.GetOptions()

	// TODO(carolynvs): this is being called twice
	_, err := p.resolveBundleReference(actionOpts)

	deperator := newDependencyExecutioner(p, action.GetAction())
	err = deperator.Prepare(action)
	if err != nil {
		return err
	}

	err = deperator.Execute()
	if err != nil {
		return err
	}

	actionArgs, err := p.BuildActionArgs(installation, action)
	if err != nil {
		return err
	}
	deperator.PrepareRootActionArguments(&actionArgs)

	fmt.Fprintf(p.Out, "%s %s...\n", action.GetActionVerb(), actionOpts.Name)
	return p.CNAB.Execute(actionArgs)
}
