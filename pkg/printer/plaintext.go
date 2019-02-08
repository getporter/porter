package printer

import (
	"fmt"
	"io"
)

// PrintPlaintext is a printer that prints the provided value as is
func PrintPlaintext(out io.Writer, v interface{}) error {
	fmt.Fprintf(out, "%v", v)
	return nil
}
