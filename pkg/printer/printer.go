package printer

import "github.com/pkg/errors"

type Format string

const (
	FormatJson      Format = "json"
	FormatTable     Format = "table"
	FormatYaml      Format = "yaml"
	FormatPlaintext Format = "plaintext"
)

func (p *PrintOptions) ParseFormat() error {
	format := Format(p.RawFormat)
	switch format {
	case FormatTable, FormatJson, FormatYaml, FormatPlaintext:
		p.Format = format
		return nil
	default:
		return errors.Errorf("invalid format: %s", p.RawFormat)
	}
}

type PrintOptions struct {
	RawFormat string
	Format
}
