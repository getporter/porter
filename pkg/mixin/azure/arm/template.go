package arm

import (
	"fmt"
	"io/ioutil"

	packr "github.com/gobuffalo/packr/v2"

	"github.com/pkg/errors"
)

// GetTemplate returns an arm template for a given service kind
// or an error if one is not found
func GetTemplate(kind string) ([]byte, error) {
	t := packr.New("templates", "./templates")
	templateFile := fmt.Sprintf("%s.template", kind)
	return t.Find(templateFile)
}

func (d deployer) FindTemplate(kind, template string) ([]byte, error) {
	if kind == "arm" {
		template := fmt.Sprintf("/cnab/arm/templates/%s", template)
		f, err := d.context.FileSystem.Open(template)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("couldn't find template %s", template))
		}
		defer f.Close()
		return ioutil.ReadAll(f)
	}
	return GetTemplate(kind)
}
