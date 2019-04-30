package porter

import (
	"github.com/radu-matei/coras/pkg/coras"
)

type PullOptions struct {
	TargetRef string
	sharedOptions
}

func (p *Porter) Pull(opts PullOptions) error {
	return coras.Pull(opts.TargetRef, opts.File, false)
}
