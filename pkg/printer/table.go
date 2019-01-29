package printer

import (
	"fmt"
	"io"
	"reflect"
	"text/tabwriter"

	"github.com/pkg/errors"
)

func NewTableWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, 0, 0, 1, ' ', tabwriter.TabIndent)
}

func PrintTable(out io.Writer, v interface{}, getRow func(row interface{}) []interface{}, headers ...interface{}) error {
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		return errors.Errorf("invalid data passed to PrintTable, must be a slice but got %T", v)
	}
	rows := reflect.ValueOf(v)

	table := NewTableWriter(out)
	if len(headers) > 0 {
		fmt.Fprintln(table, headers...)
	}
	for i := 0; i < rows.Len(); i++ {
		fmt.Fprintln(table, getRow(rows.Index(i).Interface())...)
	}
	return table.Flush()
}
