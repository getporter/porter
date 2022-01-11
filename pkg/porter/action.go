package porter

import (
	"context"

	"get.porter.sh/porter/pkg/claims"
)

// ExecuteAction runs the specified action. Supported actions are: install, upgrade, invoke.
// The uninstall action works in reverse so it's implemented separately.
func (p *Porter) ExecuteAction(ctx context.Context, installation claims.Installation, action BundleAction) error {
	deperator := newDependencyExecutioner(p, installation, action)
	err := deperator.Prepare(ctx)
	if err != nil {
		return err
	}

	err = deperator.Execute(ctx)
	if err != nil {
		return err
	}

	actionArgs, err := deperator.PrepareRootActionArguments()
	if err != nil {
		return err
	}

	return p.CNAB.Execute(ctx, actionArgs)
}
