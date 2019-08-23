package docs

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type DocsOptions struct {
	RootCommand *cobra.Command
	Destination string
}

const DefaultDestination = "./docs/content/cli/"

func (o *DocsOptions) Validate() error {
	if o.Destination == "" {
		o.Destination = DefaultDestination
	}

	return nil
}

func GenerateCliDocs(opts *DocsOptions) error {
	err := doc.GenMarkdownTreeCustom(opts.RootCommand, opts.Destination, docfileHandler(), doclinkHandler())
	if err != nil {
		return errors.Wrap(err, "error generating the markdown documentation from the cli")
	}

	// Strip off the leading porter_ from every file
	items, err := filepath.Glob(filepath.Join(opts.Destination, "porter_*.md"))
	if err != nil {
		return errors.Wrapf(err, "unable to list generated cli docs directory %q", opts.Destination)
	}

	for _, i := range items {
		inew := strings.Replace(i, "porter_", "", -1)
		err := os.Rename(i, inew)
		if err != nil {
			return errors.Wrapf(err, "unable to rename markdown file")
		}
	}
	return nil
}

func docfileHandler() func(string) string {
	const fmTemplate = `---
date: %s
title: "%s"
slug: %s
url: %s
---
`

	filePrepender := func(filename string) string {
		now := time.Now().Format(time.RFC3339)
		name := filepath.Base(filename)
		base := strings.TrimSuffix(name, path.Ext(name))
		url := "/cli/" + strings.ToLower(base) + "/"
		return fmt.Sprintf(fmTemplate, now, strings.Replace(base, "_", " ", -1), base, url)
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
