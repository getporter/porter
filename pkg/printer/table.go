package printer

import (
	"fmt"
	"io"
	"reflect"

	"github.com/olekukonko/tablewriter"
)

func NewTableSection(out io.Writer) *tablewriter.Table {
	table := tablewriter.NewWriter(out)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorders(tablewriter.Border{Left: false, Right: false, Bottom: false, Top: true})
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)
	return table
}

func PrintTableParameterSet(out io.Writer, params [][]string, headers ...string) error {
	table := NewTableSection(out)

	// Print the outputs table
	table.SetHeader(headers)
	for _, v := range params {
		table.Append(v)
	}
	table.Render()
	return nil
}

// PrintTable outputs a dataset in tabular format
func PrintTable(out io.Writer, v interface{}, getRow func(row interface{}) []string, headers ...string) error {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		return fmt.Errorf("invalid data passed to PrintTable, must be a slice but got %T", v)
	}

	rows := reflect.ValueOf(v)

	table := NewTableSection(out)

	// Print the outputs table
	table.SetHeader(headers)
	for i := 0; i < rows.Len(); i++ {
		table.Append(getRow(rows.Index(i).Interface()))
	}

	table.Render()
	return nil
}
