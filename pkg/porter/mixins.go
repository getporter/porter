package porter

import (
	"fmt"
	"path/filepath"

	"github.com/deislabs/porter/pkg/printer"
	"github.com/pkg/errors"
)

type MixinMetaData struct {
	// Mixin Name
	Name string
	// Mixin Directory
	Dir string
	// Path to the client executable
	ClientPath string
	// Version
	// Repository or Source (where did it come from)
	// Author
	// Is it up to date
	// etc
}

func (p *Porter) PrintMixins(opts printer.PrintOptions) error {
	mixins, err := p.GetMixins()
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatTable:
		printMixinRow :=
			func(v interface{}) []interface{} {
				m, ok := v.(MixinMetaData)
				if !ok {
					return nil
				}
				return []interface{}{m.Name}
			}
		return printer.PrintTable(p.Out, mixins, printMixinRow)
	case printer.FormatJson:
		return printer.PrintJson(p.Out, mixins)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) GetMixins() ([]MixinMetaData, error) {
	mixinsDir, err := p.GetMixinsDir()
	if err != nil {
		return nil, err
	}

	files, err := p.FileSystem.ReadDir(mixinsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list the contents of the mixins directory %q", mixinsDir)
	}

	mixins := make([]MixinMetaData, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		mixinDir := filepath.Join(mixinsDir, file.Name())
		mixins = append(mixins, MixinMetaData{
			Name:       file.Name(),
			ClientPath: filepath.Join(mixinDir, file.Name()),
			Dir:        mixinDir,
		})
	}

	return mixins, nil
}
