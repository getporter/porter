package docs

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type DocsOptions struct {
	RootCommand *cobra.Command
	Destination string
}

const DefaultDestination = "./docs/content/cli/"

func (o *DocsOptions) Validate(cxt *portercontext.Context) error {
	if o.Destination == "" {
		o.Destination = DefaultDestination
	}

	exists, err := cxt.FileSystem.Exists(o.Destination)
	if err != nil {
		return fmt.Errorf("error checking if --destination exists: %q: %w", o.Destination, err)
	}
	if !exists {
		return fmt.Errorf("--destination %q doesn't exist", o.Destination)
	}

	return nil
}

func GenerateCliDocs(opts *DocsOptions) error {
	opts.RootCommand.DisableAutoGenTag = true

	err := doc.GenMarkdownTreeCustom(opts.RootCommand, opts.Destination, docfileHandler(), doclinkHandler())
	if err != nil {
		return fmt.Errorf("error generating the markdown documentation from the cli: %w", err)
	}

	// Strip off the leading porter_ from every file
	items, err := filepath.Glob(filepath.Join(opts.Destination, "porter_*.md"))
	if err != nil {
		return fmt.Errorf("unable to list generated cli docs directory %q: %w", opts.Destination, err)
	}

	for _, i := range items {
		inew := strings.Replace(i, "porter_", "", -1)
		err := os.Rename(i, inew)
		if err != nil {
			return fmt.Errorf("unable to rename markdown file: %w", err)
		}
	}
	return nil
}

func docfileHandler() func(string) string {
	const fmTemplate = `---
title: "%s"
slug: %s
url: %s
---
`

	filePrepender := func(filename string) string {
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))
		url := "/cli/" + strings.ToLower(base) + "/"
		return fmt.Sprintf(fmTemplate, strings.Replace(base, "_", " ", -1), base, url)
	}
	return filePrepender
}

func doclinkHandler() func(string) string {
	linkHandler := func(name string) string {
		base := strings.TrimSuffix(name, path.Ext(name))
		return "/cli/" + strings.ToLower(base) + "/"
	}
	return linkHandler
}
