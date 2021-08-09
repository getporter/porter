package printer

import (
	"fmt"
	"io"

	"get.porter.sh/porter/pkg/encoding"
	"github.com/pkg/errors"
)

func PrintJson(out io.Writer, v interface{}) error {
	b, err := encoding.MarshalJson(v)
	if err != nil {
		return errors.Wrap(err, "could not marshal value to json")
	}
	fmt.Fprintln(out, string(b))
	return nil
}
