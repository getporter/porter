package exec

import (
	"fmt"

	packr "github.com/gobuffalo/packr/v2"
)

func (m *Mixin) PrintSchema() error {
	schema, err := m.GetSchema()
	if err != nil {
		return err
	}

	fmt.Fprintf(m.Out, schema)

	return nil
}

func (m *Mixin) GetSchema() (string, error) {
	t := packr.New("schema", "./schema")

	return t.FindString("exec.json")
}
