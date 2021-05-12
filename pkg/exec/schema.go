package exec

import (
	_ "embed"
	"fmt"
)

//go:embed schema/exec.json
var schema string

func (m *Mixin) PrintSchema() {
	fmt.Fprintf(m.Out, m.GetSchema())
}

func (m *Mixin) GetSchema() string {
	return schema
}
