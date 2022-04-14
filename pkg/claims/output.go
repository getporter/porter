package claims

import (
	"sort"

	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/secrets"
	"get.porter.sh/porter/pkg/storage"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/schema"
	"github.com/pkg/errors"
)

var _ storage.Document = Output{}

type Output struct {
	SchemaVersion schema.Version `json:"schemaVersion" yaml:"schemaVersion" toml:"schemaVersion"`
	Name          string         `json:"name" yaml:"name" toml:"name"`
	Namespace     string         `json:"namespace" yaml:"namespace" toml:"namespace"`
	Installation  string         `json:"installation" yaml:"installation" toml:"installation"`
	RunID         string         `json:"runId" yaml:"runId" toml:"runId"`
	ResultID      string         `json:"resultId" yaml:"resultId" toml:"resultId"`
	Value         []byte         `json:"value" yaml:"value" toml:"value"`
}

func (o Output) DefaultDocumentFilter() interface{} {
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

// FormatSensitive transforms an output value to a secret key that can be used
// to store/retrieve the sensitive output value from a secret store.
// It returns the transformed output.
func (o Output) FormatSensitive() Output {
	o.Value = []byte(o.RunID + o.Name)
	return o
}

func (o Output) Resolve(store secrets.Store) (Output, error) {
	resolved, err := store.Resolve(secrets.SourceSecret, string(o.Value))
	if err != nil {
		return o, err
	}

	o.Value = []byte(resolved)
	return o, nil
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

func (o Outputs) Resolve(store secrets.Store, bun cnab.ExtendedBundle) (Outputs, error) {
	for name, idx := range o.keys {
		output, ok := o.GetByIndex(idx)
		if !ok {
			continue
		}

		sensitive, err := bun.IsOutputSensitive(name)
		if err != nil || !sensitive {
			o.vals[idx] = output
			continue
		}

		resolved, err := output.Resolve(store)
		if err != nil {
			return o, errors.WithMessagef(err, "failed to resolve output %q", output.Name)
		}
		o.vals[idx] = resolved
	}

	return o, nil
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
