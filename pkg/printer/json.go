package printer

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func PrintJson(out io.Writer, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return errors.Wrap(err, "could not marshal value to json")
	}
	fmt.Fprintln(out, string(b))
	return nil
}
