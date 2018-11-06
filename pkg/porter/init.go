package porter

import "fmt"

func (p *Porter) Init() error {
	fmt.Fprintln(p.Out, "initializing porter configuration...")
	return nil
}
