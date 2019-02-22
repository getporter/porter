package porter

import (
	"fmt"

	"github.com/deislabs/porter/pkg/mixin"

	"github.com/deislabs/porter/pkg/printer"
)

// MixinProvider handles searching, listing and communicating with the mixins.
type MixinProvider interface {
	GetMixins() ([]mixin.Metadata, error)
	GetMixinSchema(m mixin.Metadata) (map[string]interface{}, error)
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
				m, ok := v.(mixin.Metadata)
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
