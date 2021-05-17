package porter

import "get.porter.sh/porter/pkg/context"

type contextOptions struct {
	Verbose bool
}

func NewContextOptions(cxt *context.Context) contextOptions {
	return contextOptions{
		Verbose: cxt.IsVerbose(),
	}
}

func (o contextOptions) Apply(cxt *context.Context) {
	cxt.SetVerbose(o.Verbose)
}
