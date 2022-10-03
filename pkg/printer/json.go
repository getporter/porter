package printer

import (
	"fmt"
	"io"

	"get.porter.sh/porter/pkg/encoding"
)

func PrintJson(out io.Writer, v interface{}) error {
	b, err := encoding.MarshalJson(v)
	if err != nil {
		return fmt.Errorf("could not marshal value to json: %w", err)
	}
	fmt.Fprintln(out, string(b))
	return nil
}
