package feed

import (
	_ "embed"
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/pkg/errors"
)

//go:embed templates/atom-template.xml
var feedTemplate []byte

func CreateTemplate(cxt *portercontext.Context) error {
	templateFile := "atom-template.xml"
	err := cxt.FileSystem.WriteFile(templateFile, feedTemplate, 0644)
	if err != nil {
		return errors.Wrapf(err, "error writing mixin feed template to %s", templateFile)
	}

	fmt.Fprintf(cxt.Out, "wrote mixin feed template to %s\n", templateFile)
	return nil
}
