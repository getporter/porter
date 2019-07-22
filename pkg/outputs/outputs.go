package outputs

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/config"
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
func ReadBundleOutput(c *config.Config, name, claim string) (*Output, error) {
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
