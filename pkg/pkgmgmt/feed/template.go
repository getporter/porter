package feed

import (
	"fmt"

	"get.porter.sh/porter/pkg/context"
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
)

func CreateTemplate(cxt *context.Context) error {
	box := NewTemplatesBox()
	tmpl, err := box.Find("atom-template.xml")
	if err != nil {
		return errors.Wrap(err, "error loading mixin feed template")
	}

	templateFile := "atom-template.xml"
	err = cxt.FileSystem.WriteFile(templateFile, tmpl, 0644)
	if err != nil {
		return errors.Wrapf(err, "error writing mixin feed template to %s", templateFile)
	}

	fmt.Fprintf(cxt.Out, "wrote mixin feed template to %s\n", templateFile)
	return nil
}

// NewSchemas creates or retrieves the packr box with the porter template files.
func NewTemplatesBox() *packr.Box {
	return packr.New("get.porter.sh/porter/pkg/mixin/feed/templates", "./templates")
}
