package helm

import (
	"fmt"
)

func (m *Mixin) Build() error {
	fmt.Fprintln(m.Out, "RUN echo 'TODO: COPY HELM BINARY'")
	return nil
}
