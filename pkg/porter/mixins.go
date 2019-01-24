package porter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/tabwriter"

	"github.com/pkg/errors"
)

type MixinMetaData struct {
	Name string
	// Version
	// Repository or Source (where did it come from)
	// Author
	// Is it up to date
	// etc
}

func (p *Porter) PrintMixins(format string) error {
	mixins, err := p.GetMixins()
	if err != nil {
		return err
	}

	switch format {
	case "human":
		table := tabwriter.NewWriter(p.Out, 0, 0, 1, ' ', tabwriter.TabIndent)
		fmt.Fprintln(table, "Name")
		for _, mixin := range mixins {
			fmt.Fprintln(table, mixin.Name)
		}
		table.Flush()

	case "json":
		b, err := json.MarshalIndent(mixins, "", "  ")
		if err != nil {
			return errors.Wrap(err, "could not marshal mixins to json")
		}
		fmt.Fprintln(p.Out, string(b))
	}
	return nil
}

func (p *Porter) GetMixins() ([]MixinMetaData, error) {
	mixinsDir, err := p.GetMixinsDir()
	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(mixinsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list the contents of the mixins directory %q", mixinsDir)
	}

	mixins := make([]MixinMetaData, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		mixins = append(mixins, MixinMetaData{Name: file.Name()})
	}

	return mixins, nil
}
