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
		fmt.Fprintln(table, tabify(headers)...)
	}
	for i := 0; i < rows.Len(); i++ {
		fmt.Fprintln(table, tabify(getRow(rows.Index(i).Interface()))...)
	}
	return table.Flush()
}

// tabify is a helper function which takes a slice and injects tab characters
// between each element such that tabwriter can work its magic
func tabify(untabified []interface{}) []interface{} {
	var tabified []interface{}
	for i := 0; i < len(untabified); i++ {
		tabified = append(tabified, untabified[i])
		// only append tab character if prior to last element
		if i+1 < len(untabified) {
			tabified = append(tabified, "\t")
		}
	}
	return tabified
}
