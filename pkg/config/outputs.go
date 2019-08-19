package config

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/pkg/errors"
)

// Output represents a bundle output
type Output struct {
	Name      string `json:"name"`
	Sensitive bool   `json:"sensitive"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

// Outputs is a slice of Outputs
type Outputs []Output

func (l Outputs) Len() int {
	return len(l)
}
func (l Outputs) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (l Outputs) Less(i, j int) bool {
	var si = l[i].Name
	var sj = l[j].Name
	var siLower = strings.ToLower(si)
	var sjLower = strings.ToLower(sj)
	if siLower == sjLower {
		return si < sj
	}
	return siLower < sjLower
}

// ReadBundleOutput reads the provided output associated with the provided bundle,
// via the filesystem provided by the config.Config object,
// returning the output's full Output representation
func (c *Config) ReadBundleOutput(name string, claim string) (*Output, error) {
	outputsDir, err := c.GetOutputsDir()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get outputs directory")
	}
	bundleOutputsDir := filepath.Join(outputsDir, claim)

	outputPath := filepath.Join(bundleOutputsDir, name)

	bytes, err := c.Context.FileSystem.ReadFile(outputPath)
	if err != nil {
		return nil, errors.Errorf("unable to read output '%s' for claim '%s'", name, claim)
	}

	var output Output
	err = json.Unmarshal(bytes, &output)
	if err != nil {
		return nil, errors.Errorf("unable to unmarshal output '%s' for claim '%s'", name, claim)
	}

	return &output, nil
}

// JSONMarshal marshals an Output to JSON, returning a byte array or error
func (o *Output) JSONMarshal() ([]byte, error) {
	return json.MarshalIndent(o, "", "  ")
}

// TODO: remove in favor of cnab-go logic: https://github.com/deislabs/cnab-go/pull/99
func OutputAppliesTo(action string, output bundle.Output) bool {
	if len(output.ApplyTo) == 0 {
		return true
	}
	for _, act := range output.ApplyTo {
		if action == act {
			return true
		}
	}
	return false
}
