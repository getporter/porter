package porter

import "get.porter.sh/porter/pkg/portercontext"

type contextOptions struct {
	Verbose bool
}

func NewContextOptions(cxt *portercontext.Context) contextOptions {
	return contextOptions{
		Verbose: cxt.IsVerbose(),
	}
}

func (o contextOptions) Apply(cxt *portercontext.Context) {
	cxt.SetVerbose(o.Verbose)
}
