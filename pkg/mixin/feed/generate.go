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

type mixinFileset struct {
	Mixin   string
	Version string
	Files   []mixinFile
}

func (f *mixinFileset) Updated() string {
	return toAtomTimestamp(f.GetLastUpdated())
}

func (f *mixinFileset) GetLastUpdated() time.Time {
	var max time.Time
	for _, f := range f.Files {
		if f.Updated.After(max) {
			max = f.Updated
		}
	}
	return max
}

type mixinEntries []*mixinFileset

func (e mixinEntries) Len() int {
	return len(e)
}

func (e mixinEntries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e mixinEntries) Less(i, j int) bool {
	return e[i].GetLastUpdated().Before(e[j].GetLastUpdated())
}

type mixinFile struct {
	File    string
	Updated time.Time
}

type mixinFeed map[string]map[string]*mixinFileset

func Generate(opts GenerateOptions, cxt *context.Context) error {
	feedTmpl, err := cxt.FileSystem.ReadFile(opts.TemplateFile)
	if err != nil {
		return errors.Wrapf(err, "error reading template file at %s", opts.TemplateFile)
	}

	mixinRegex := regexp.MustCompile(`(.*/)?(.+)/([a-z]+)-(linux|windows|darwin)-(amd64)(\.exe)?`)

	feed := mixinFeed{}
	found := 0
	err = cxt.FileSystem.Walk(opts.SearchDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		matches := mixinRegex.FindStringSubmatch(path)
		if len(matches) > 0 {
			version := matches[2]
			mixin := matches[3]
			filename := info.Name()

			versions, ok := feed[mixin]
			if !ok {
				versions = map[string]*mixinFileset{}
				feed[mixin] = versions
			}

			fileset, ok := versions[version]
			if !ok {
				fileset = &mixinFileset{
					Mixin:   mixin,
					Version: version,
				}
				versions[version] = fileset
				found++
			}
			fileset.Files = append(fileset.Files, mixinFile{File: filename, Updated: info.ModTime()})
		}

		return nil
	})
	if err != nil {
		return err
	}

	tmplData := map[string]interface{}{}
	mixins := make([]string, 0, len(feed))
	entries := make(mixinEntries, 0, found)
	for m, versions := range feed {
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
	err = cxt.FileSystem.WriteFile(opts.AtomFile, []byte(atomXml), 0644)
	if err != nil {
		return errors.Wrapf(err, "could not write feed to %s", opts.AtomFile)
	}

	fmt.Fprintf(cxt.Out, "wrote feed to %s\n", opts.AtomFile)
	return nil
}

func toAtomTimestamp(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
