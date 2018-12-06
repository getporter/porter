package arm

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"
)

// Render renders the given a Go text template as a []byte and any object to be
// exposed within the template as "." into the resulting []byte. It also permits
// the use of functions from the sprig library.
func Render(tpl []byte, obj interface{}) ([]byte, error) {
	goTemplate, err := template.New(
		"",
	).Funcs(sprig.TxtFuncMap()).Parse(string(tpl))
	if err != nil {
		return nil, fmt.Errorf("error creating Go template: %s", err)
	}
	var armTemplateBuffer bytes.Buffer
	err = goTemplate.Execute(&armTemplateBuffer, obj)
	if err != nil {
		return nil, fmt.Errorf(
			"error rendering Go template into ARM template: %s",
			err,
		)
	}
	return armTemplateBuffer.Bytes(), nil
}
