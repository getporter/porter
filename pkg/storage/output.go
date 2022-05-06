package storage

import (
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/schema"
)

var _ Document = Output{}

type Output struct {
	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`
	ID            string         `json:"_id" yaml:"_id" toml:"_id"`
	Name          string         `json:"name" yaml:"name" toml:"name"`
	Namespace     string         `json:"namespace" yaml:"namespace" toml:"namespace"`
	Installation  string         `json:"installation" yaml:"installation" toml:"installation"`
	RunID         string         `json:"runId" yaml:"runId" toml:"runId"`
	ResultID      string         `json:"resultId" yaml:"resultId" toml:"resultId"`

	// Key holds the secret key to retrieve a sensitive output value
	Key   string `json:"key" yaml:"key" toml:"key"`
	Value []byte `json:"value" yaml:"value" toml:"value"`
}

func (o Output) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"resultId": o.ResultID, "name": o.Name}
}

// GetSchema returns the schema for the output from the specified bundle, or
// false if the schema is not defined.
func (o Output) GetSchema(b cnab.ExtendedBundle) (definition.Schema, bool) {
	if def, ok := b.Outputs[o.Name]; ok {
		if schema, ok := b.Definitions[def.Definition]; ok {
			return *schema, ok
		}
	}

	return definition.Schema{}, false
}

type Outputs struct {
	// Sorted list of outputs
	vals []Output
	// output name -> index of the output in vals
	keys map[string]int
}

func NewOutputs(outputs []Output) Outputs {
	o := Outputs{
		vals: make([]Output, len(outputs)),
		keys: make(map[string]int, len(outputs)),
	}

	copy(o.vals, outputs)
	for i, output := range outputs {
		o.keys[output.Name] = i
	}

	sort.Sort(o)
	return o
}

func (o Outputs) GetByName(name string) (Output, bool) {
	i, ok := o.keys[name]
	if !ok || i >= len(o.vals) {
		return Output{}, false
	}

	return o.vals[i], true
}

func (o Outputs) GetByIndex(i int) (Output, bool) {
	if i < 0 || i >= len(o.vals) {
		return Output{}, false
	}

	return o.vals[i], true
}

// Value returns a list of outputs.
func (o Outputs) Value() []Output {
	return o.vals
}

func (o Outputs) Len() int {
	return len(o.vals)
}

func (o Outputs) Less(i, j int) bool {
	return o.vals[i].Name < o.vals[j].Name
}

func (o Outputs) Swap(i, j int) {
	o.keys[o.vals[i].Name] = j
	o.keys[o.vals[j].Name] = i
	o.vals[i], o.vals[j] = o.vals[j], o.vals[i]
}
