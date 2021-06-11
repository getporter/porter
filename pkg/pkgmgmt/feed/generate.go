package feed

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/context"
	"github.com/Masterminds/semver/v3"
	"github.com/cbroglie/mustache"
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
		o.SearchDirectory = cxt.Getwd()
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
	existingFeed, err := feed.FileSystem.Exists(opts.AtomFile)
	if err != nil {
		return err
	}
	if existingFeed {
		err := feed.Load(opts.AtomFile)
		if err != nil {
			return err
		}
	}

	mixinRegex := regexp.MustCompile(`(.*/)?(.+)/([a-z0-9-]+)-(linux|windows|darwin)-(amd64)(\.exe)?`)

	err = feed.FileSystem.Walk(opts.SearchDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		matches := mixinRegex.FindStringSubmatch(path)
		if len(matches) > 0 {
			version := matches[2]

			if !shouldPublishVersion(version) {
				return nil
			}

			mixin := matches[3]
			filename := info.Name()
			updated := info.ModTime()

			_, ok := feed.Index[mixin]
			if !ok {
				feed.Index[mixin] = map[string]*MixinFileset{}
			}

			_, ok = feed.Index[mixin][version]
			if !ok {
				fileset := MixinFileset{
					Mixin:   mixin,
					Version: version,
				}

				feed.Index[mixin][version] = &fileset
			}

			for i := range feed.Index[mixin][version].Files {
				mixinFile := feed.Index[mixin][version].Files[i]
				if mixinFile.File == filename {
					if mixinFile.Updated.Before(updated) {
						mixinFile.Updated = updated
					}

					return nil
				}
			}

			feed.Index[mixin][version].Files = append(feed.Index[mixin][version].Files, &MixinFile{File: filename, Updated: updated})
		}

		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "failed to traverse the %s directory", opts.SearchDirectory)
	}

	if len(feed.Index) == 0 {
		return fmt.Errorf("no mixin binaries found in %s matching the regex %q", opts.SearchDirectory, mixinRegex)
	}

	return nil
}

var versionRegex = regexp.MustCompile(`\d+-g[a-z0-9]+`)

// As a safety measure, skip versions that shouldn't be put in the feed, we only want canary and tagged releases.
func shouldPublishVersion(version string) bool {
	if strings.HasSuffix(version, "canary") {
		// Publish canary permalinks
		return true
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		// If it's not a version, don't publish
		return false
	}

	// Check if this is an untagged version, i.e. the output of git describe, v1.2.3-2-ga1b3c5
	untagged := versionRegex.MatchString(v.Prerelease())
	return !untagged
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

	sort.Strings(mixins)

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
