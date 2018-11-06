package porter

import "fmt"

func (p *Porter) Build() error {
	fmt.Fprintln(p.Out, "generating CNAB bundle...")
	return nil
}
