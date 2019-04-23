package feed

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/cbroglie/mustache"
	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

type GenerateOptions struct {
	SearchDirectory string
	AtomFile        string
	TemplateFile    string
}

func (o *GenerateOptions) Validate(c *context.Context) error {
	err := o.ValidateSearchDirectory(c)
	if err != nil {
		return err
	}

	return o.ValidateTemplateFile(c)
}

func (o *GenerateOptions) ValidateSearchDirectory(cxt *context.Context) error {
	if o.SearchDirectory == "" {
		wd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "could not get current working directory")
		}

		o.SearchDirectory = wd
	}

	if _, err := cxt.FileSystem.Stat(o.SearchDirectory); err != nil {
		return errors.Wrapf(err, "invalid --dir %s", o.SearchDirectory)
	}

	return nil
}

func (o *GenerateOptions) ValidateTemplateFile(cxt *context.Context) error {
	if _, err := cxt.FileSystem.Stat(o.TemplateFile); err != nil {
		return errors.Wrapf(err, "invalid --template %s", o.TemplateFile)
	}

	return nil
}

func (feed *MixinFeed) Generate(opts GenerateOptions) error {
	mixinRegex := regexp.MustCompile(`(.*/)?(.+)/([a-z]+)-(linux|windows|darwin)-(amd64)(\.exe)?`)

	feed.Index = make(map[string]map[string]*MixinFileset)
	return feed.FileSystem.Walk(opts.SearchDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		matches := mixinRegex.FindStringSubmatch(path)
		if len(matches) > 0 {
			version := matches[2]
			mixin := matches[3]
			filename := info.Name()

			versions, ok := feed.Index[mixin]
			if !ok {
				versions = map[string]*MixinFileset{}
				feed.Index[mixin] = versions
			}

			fileset, ok := versions[version]
			if !ok {
				fileset = &MixinFileset{
					Mixin:   mixin,
					Version: version,
				}
				versions[version] = fileset
			}
			fileset.Files = append(fileset.Files, MixinFile{File: filename, Updated: info.ModTime()})
		}

		return nil
	})
}

func (feed *MixinFeed) Save(opts GenerateOptions) error {
	feedTmpl, err := feed.FileSystem.ReadFile(opts.TemplateFile)
	if err != nil {
		return errors.Wrapf(err, "error reading template file at %s", opts.TemplateFile)
	}

	tmplData := map[string]interface{}{}
	mixins := make([]string, 0, len(feed.Index))
	entries := make(MixinEntries, 0, len(feed.Index))
	for m, versions := range feed.Index {
		mixins = append(mixins, m)
		for _, fileset := range versions {
			entries = append(entries, fileset)
		}
	}
	sort.Sort(sort.Reverse(entries))

	tmplData["Mixins"] = mixins
	tmplData["Entries"] = entries
	tmplData["Updated"] = entries[0].Updated()

	atomXml, err := mustache.Render(string(feedTmpl), tmplData)
	err = feed.FileSystem.WriteFile(opts.AtomFile, []byte(atomXml), 0644)
	if err != nil {
		return errors.Wrapf(err, "could not write feed to %s", opts.AtomFile)
	}

	fmt.Fprintf(feed.Out, "wrote feed to %s\n", opts.AtomFile)
	return nil
}

func toAtomTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
