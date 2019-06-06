package porter

import "github.com/deislabs/porter/pkg/context"

type contextOptions struct {
	Verbose bool
}

func (o contextOptions) Apply(cxt *context.Context) {
	cxt.SetVerbose(o.Verbose)
}
