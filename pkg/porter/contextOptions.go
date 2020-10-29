package porter

import "get.porter.sh/porter/pkg/context"

type contextOptions struct {
	Verbose bool
}

func (o contextOptions) Apply(cxt *context.Context) {
	cxt.SetVerbose(o.Verbose)
}
