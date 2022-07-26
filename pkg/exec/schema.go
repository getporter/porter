package exec

import (
	_ "embed"
	"fmt"
)

//go:embed schema/exec.json
var schema string

func (m *Mixin) PrintSchema() {
	fmt.Fprint(m.Config.Out, m.GetSchema())
}

func (m *Mixin) GetSchema() string {
	return schema
}
