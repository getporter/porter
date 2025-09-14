package printer

import (
	"fmt"
	"io"
	"reflect"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func NewTableSection(out io.Writer) *tablewriter.Table {
	table := tablewriter.NewTable(out,
		tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{
			Settings: tw.Settings{
				Separators: tw.Separators{
					BetweenColumns: tw.Off,
				},
			},
			Borders: tw.Border{
				Top:    tw.On,
				Left:   tw.Off,
				Right:  tw.Off,
				Bottom: tw.Off,
			},
		})),
		tablewriter.WithConfig(tablewriter.Config{
			Header: tw.CellConfig{
				Formatting: tw.CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoFormat: tw.Off,
				},
			},
			Row: tw.CellConfig{
				ColMaxWidths: tw.CellWidth{Global: 30},
				Formatting: tw.CellFormatting{
					Alignment:  tw.AlignLeft,
					AutoWrap:   tw.WrapNormal,
					AutoFormat: tw.Off,
				},
			},
		}),
	)

	return table
}

func PrintTableParameterSet(out io.Writer, params [][]string, headers ...string) error {
	table := NewTableSection(out)

	// Print the outputs table
	table.Header(headers)
	for _, v := range params {
		err := table.Append(v)
		if err != nil {
			return err
		}
	}
	return table.Render()
}

// PrintTable outputs a dataset in tabular format
func PrintTable(out io.Writer, v interface{}, getRow func(row interface{}) []string, headers ...string) error {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		return fmt.Errorf("invalid data passed to PrintTable, must be a slice but got %T", v)
	}

	rows := reflect.ValueOf(v)

	table := NewTableSection(out)

	// Print the outputs table
	table.Header(headers)
	for i := 0; i < rows.Len(); i++ {
		err := table.Append(getRow(rows.Index(i).Interface()))
		if err != nil {
			return err
		}

	}

	return table.Render()
}
