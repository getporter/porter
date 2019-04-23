package feed

import (
	"bytes"
	"fmt"
	"net/url"

	"github.com/deislabs/porter/pkg/context"
	"github.com/mmcdole/gofeed/atom"
	"github.com/pkg/errors"
)

func (feed *MixinFeed) Load(file string, cxt *context.Context) error {
	contents, err := cxt.FileSystem.ReadFile(file)
	if err != nil {
		return errors.Wrapf(err, "error reading mixin feed at %s", file)
	}

	p := atom.Parser{}
	atomFeed, err := p.Parse(bytes.NewReader(contents))
	if err != nil {
		if cxt.Debug {
			fmt.Fprintln(cxt.Err, string(contents))
		}
		return errors.Wrap(err, "error parsing the mixin feed as an atom xml file")
	}

	feed.Updated = atomFeed.UpdatedParsed

	for _, category := range atomFeed.Categories {
		feed.Mixins = append(feed.Mixins, category.Term)
	}

	feed.Index = make(map[string]map[string]*MixinFileset)
	for _, entry := range atomFeed.Entries {
		fileset := &MixinFileset{}

		if len(entry.Categories) == 0 {
			if cxt.Debug {
				fmt.Fprintf(cxt.Err, "skipping invalid entry %s, missing category (mixin name)", entry.ID)
			}
			continue
		}
		fileset.Mixin = entry.Categories[0].Term
		if fileset.Mixin == "" {
			if cxt.Debug {
				fmt.Fprintf(cxt.Err, "skipping invalid entry %s, empty category (mixin name)", entry.ID)
			}
			continue
		}

		fileset.Version = entry.Content.Value
		if fileset.Version == "" {
			if cxt.Debug {
				fmt.Fprintf(cxt.Err, "skipping invalid entry %s, empty content (version)", entry.ID)
			}
			continue
		}

		fileset.Files = make([]MixinFile, 0, len(entry.Links))
		for _, link := range entry.Links {
			if link.Rel == "download" {
				if entry.UpdatedParsed == nil {
					if cxt.Debug {
						fmt.Fprintf(cxt.Err, "skipping invalid entry %s, invalid updated %q could not be parsed as RFC3339", entry.ID, entry.Updated)
					}
					continue
				}

				parsedUrl, err := url.Parse(link.Href)
				if err != nil || link.Href == "" {
					if cxt.Debug {
						fmt.Fprintf(cxt.Err, "skipping invalid entry %s, invalid link.href %q", entry.ID, link.Href)
					}
					continue
				}

				file := MixinFile{
					URL:     parsedUrl,
					Updated: *entry.UpdatedParsed,
				}
				fileset.Files = append(fileset.Files, file)
			}
		}
		versions, ok := feed.Index[fileset.Mixin]
		if !ok {
			versions = map[string]*MixinFileset{}
			feed.Index[fileset.Mixin] = versions
		}

		indexedFileset, ok := versions[fileset.Version]
		if !ok || fileset.GetLastUpdated().After(indexedFileset.GetLastUpdated()) {
			versions[fileset.Version] = fileset
		}
	}

	return nil
}
