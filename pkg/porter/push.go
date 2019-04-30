package porter

import "github.com/radu-matei/coras/pkg/coras"

type PushOptions struct {
	TargetRef string
	sharedOptions
}

// Push pushes a build CNAB bundle to an OCI registry
func (p *Porter) Push(opts PushOptions) error {
	err := p.applyDefaultOptions(&opts.sharedOptions)
	if err != nil {
		return err
	}

	return coras.Push(opts.File, opts.TargetRef, false)
}
