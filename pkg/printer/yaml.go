package printer

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
)

// PrintYaml is a printer that prints the provided value in yaml
func PrintYaml(out io.Writer, v interface{}) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "could not marshal value to yaml")
	}
	fmt.Fprintln(out, string(b))
	return nil
}
