package printer

import "github.com/pkg/errors"

type Format string

const (
	FormatJson      Format = "json"
	FormatTable     Format = "table"
	FormatYaml      Format = "yaml"
	FormatPlaintext Format = "plaintext"
)

func ParseFormat(v string) (Format, error) {
	format := Format(v)
	switch format {
	case FormatTable, FormatJson, FormatYaml, FormatPlaintext:
		return format, nil
	default:
		return "", errors.Errorf("invalid format: %s", v)
	}
}

type PrintOptions struct {
	Format
}
