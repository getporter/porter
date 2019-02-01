package printer

import (
	"fmt"
	"io"
)

// PrintRaw is a printer that prints the provided value as is
func PrintRaw(out io.Writer, v interface{}) error {
	fmt.Fprintln(out, v)
	return nil
}
