package kubernetes

import (
	"fmt"

	"github.com/deislabs/porter/pkg"
)

func (m *Mixin) PrintVersion() {
	fmt.Fprintf(m.Out, "kubernetes mixin %s (%s)\n", pkg.Version, pkg.Commit)
}
