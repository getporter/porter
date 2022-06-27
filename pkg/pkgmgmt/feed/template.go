package feed

import (
	_ "embed"
	"fmt"

	"get.porter.sh/porter/pkg/portercontext"
)

//go:embed templates/atom-template.xml
var feedTemplate []byte

func CreateTemplate(cxt *portercontext.Context) error {
	templateFile := "atom-template.xml"
	err := cxt.FileSystem.WriteFile(templateFile, feedTemplate, 0644)
	if err != nil {
		return fmt.Errorf("error writing mixin feed template to %s: %w", templateFile, err)
	}

	fmt.Fprintf(cxt.Out, "wrote mixin feed template to %s\n", templateFile)
	return nil
}
