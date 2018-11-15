package exec

import (
	"fmt"

	"github.com/deislabs/porter/pkg"
)

func (m *Mixin) PrintVersion() {
	fmt.Fprintf(m.Out, "exec mixin %s (%s)\n", pkg.Version, pkg.Commit)
}
