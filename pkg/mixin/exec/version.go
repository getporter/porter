package exec

import (
	"fmt"

	"github.com/deislabs/porter/pkg"
)

func (e *Exec) PrintVersion() {
	fmt.Fprintf(e.Out, "exec mixin %s (%s)\n", pkg.Version, pkg.Commit)
}
