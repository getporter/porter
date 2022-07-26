package feed

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"path"

	"get.porter.sh/porter/pkg/tracing"
	"github.com/mmcdole/gofeed/atom"
	"go.opentelemetry.io/otel/attribute"
)

func (feed *MixinFeed) Load(ctx context.Context, file string) error {
	//lint:ignore SA4006 ignore unused ctx for now
	ctx, span := tracing.StartSpan(ctx)
	defer span.EndSpan()

	contents, err := feed.FileSystem.ReadFile(file)
	if err != nil {
		return span.Error(fmt.Errorf("error reading mixin feed at %s: %w", file, err))
	}

	p := atom.Parser{}
	atomFeed, err := p.Parse(bytes.NewReader(contents))
	if err != nil {
		return span.Error(fmt.Errorf("error parsing the mixin feed as an atom xml file: %w", err),
			attribute.String("contents", string(contents)))
	}

	feed.Updated = atomFeed.UpdatedParsed

	for _, category := range atomFeed.Categories {
		feed.Mixins = append(feed.Mixins, category.Term)
	}

	for _, entry := range atomFeed.Entries {
		fileset := &MixinFileset{}

		if len(entry.Categories) == 0 {
			span.Debugf("skipping invalid entry %s, missing category (mixin name)", entry.ID)
			continue
		}
		fileset.Mixin = entry.Categories[0].Term
		if fileset.Mixin == "" {
			span.Debugf("skipping invalid entry %s, empty category (mixin name)", entry.ID)
			continue
		}

		fileset.Version = entry.Content.Value
		if fileset.Version == "" {
			span.Debugf("skipping invalid entry %s, empty content (version)", entry.ID)
			continue
		}

		fileset.Files = make([]*MixinFile, 0, len(entry.Links))
		for _, link := range entry.Links {
			if link.Rel == "download" {
				if entry.UpdatedParsed == nil {
					span.Debugf("skipping invalid entry %s, invalid updated %q could not be parsed as RFC3339", entry.ID, entry.Updated)
					continue
				}

				parsedUrl, err := url.Parse(link.Href)
				if err != nil || link.Href == "" {
					span.Debugf("skipping invalid entry %s, invalid link.href %q", entry.ID, link.Href)
					continue
				}

				file := &MixinFile{
					URL:     parsedUrl,
					Updated: *entry.UpdatedParsed,
					File:    path.Base(parsedUrl.Path),
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
