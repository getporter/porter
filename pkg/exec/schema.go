package exec

import (
	"fmt"
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
	return m.schemas.FindString("exec.json")
}
