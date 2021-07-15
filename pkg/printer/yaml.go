package printer

import (
	"fmt"
	"io"

	"get.porter.sh/porter/pkg/encoding"
	"github.com/pkg/errors"
)

// PrintYaml is a printer that prints the provided value in yaml
func PrintYaml(out io.Writer, v interface{}) error {
	b, err := encoding.MarshalYaml(v)
	if err != nil {
		return errors.Wrap(err, "could not marshal value to yaml")
	}
	fmt.Fprintf(out, string(b)) // yaml already includes a trailing newline, so don't print another
	return nil
}
