package helm

import (
	"fmt"
)

func (m *Mixin) Uninstall() error {
	fmt.Fprintln(m.Out, "helm uninstall not implemented")
	return nil
}
