package porter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

func (p *Porter) listBundleOutputs(bundle string) (*Outputs, error) {
	outputsDir, err := p.Config.GetOutputsDir()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get outputs directory")
	}
	bundleOutputsDir := filepath.Join(outputsDir, bundle)

	var outputList Outputs
	// Walk through bundleOutputsDir, if exists, and read all output filenames.
	// We truncate actual output values, intending for the full values to be
	// retrieved by another command.
	if ok, _ := p.Context.FileSystem.DirExists(bundleOutputsDir); ok {
		// TODO: test sad path/err here
		err := p.Context.FileSystem.Walk(bundleOutputsDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				outputData, err := p.readBundleOutput(info.Name(), bundle)
				if err != nil {
					return errors.New("unable to read output")
				}

				var output Output
				err = json.Unmarshal(outputData, &output)
				if err != nil {
					return err
				}

				outputList = append(outputList, output)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		sort.Sort(sort.Reverse(outputList))
	}
	return &outputList, nil
}

func (p *Porter) readBundleOutput(output, bundle string) ([]byte, error) {
	outputsDir, err := p.Config.GetOutputsDir()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get outputs directory")
	}
	bundleOutputsDir := filepath.Join(outputsDir, bundle)

	outputPath := filepath.Join(bundleOutputsDir, output)

	return p.Context.FileSystem.ReadFile(outputPath)
}

// TODO: refactor to truncate in the middle?  (Handy if paths are long)
func truncateString(str string, num int) string {
	truncated := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		truncated = str[0:num] + "..."
	}
	return truncated
}
