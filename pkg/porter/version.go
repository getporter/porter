package porter

import (
	"fmt"

	"github.com/deislabs/porter/pkg"
)

func (p *Porter) PrintVersion() {
	fmt.Fprintf(p.Out, "porter %s (%s)\n", pkg.Version, pkg.Commit)
}
