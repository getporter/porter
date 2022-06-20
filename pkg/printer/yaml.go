package printer

import (
	"fmt"
	"io"

	"get.porter.sh/porter/pkg/encoding"
)

// PrintYaml is a printer that prints the provided value in yaml
func PrintYaml(out io.Writer, v interface{}) error {
	b, err := encoding.MarshalYaml(v)
	if err != nil {
		return fmt.Errorf("could not marshal value to yaml: %w", err)
	}
	fmt.Fprintf(out, string(b)) // yaml already includes a trailing newline, so don't print another
	return nil
}
