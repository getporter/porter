package feed

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/Masterminds/semver/v3"
)

type MixinFeed struct {
	*portercontext.Context

	// Index of mixin files
	Index map[string]map[string]*MixinFileset

	// Mixins present in the feed
	Mixins []string

	// Updated timestamp according to the atom xml feed
	Updated *time.Time
}

func NewMixinFeed(cxt *portercontext.Context) *MixinFeed {
	return &MixinFeed{
		Index:   make(map[string]map[string]*MixinFileset),
		Context: cxt,
	}
}

func (feed *MixinFeed) Search(mixin string, version string) *MixinFileset {
	versions, ok := feed.Index[mixin]
	if !ok {
		return nil
	}

	fileset, ok := versions[version]
	if ok {
		return fileset
	}

	// Return the highest version of the requested mixin according to semver, ignoring pre-releases
	if version == "latest" {
		var latestVersion *semver.Version
		for version := range versions {
			v, err := semver.NewVersion(version)
			if err != nil {
				continue
			}
			if v.Prerelease() != "" {
				continue
			}
			if latestVersion == nil || v.GreaterThan(latestVersion) {
				latestVersion = v
			}
		}
		if latestVersion != nil {
			return versions[latestVersion.Original()]
		}
	}

	return nil
}

type MixinFileset struct {
	Mixin   string
	Version string
	Files   []*MixinFile
}

func (f *MixinFileset) FindDownloadURL(ctx context.Context, os string, arch string) *url.URL {
	log := tracing.LoggerFromContext(ctx)

	match := fmt.Sprintf("%s-%s-%s", f.Mixin, os, arch)
	for _, file := range f.Files {
		if strings.Contains(file.URL.Path, match) {
			return file.URL
		}
	}

	// Until we have full support for M1 chipsets, rely on rossetta functionality in macos and use the amd64 binary
	if os == "darwin" && arch == "arm64" {
		log.Debugf("%s @ %s did not publish a download for darwin/arm64, falling back to darwin/amd64", f.Mixin, f.Version)
		return f.FindDownloadURL(ctx, "darwin", "amd64")
	}

	return nil
}

func (f *MixinFileset) Updated() string {
	return toAtomTimestamp(f.GetLastUpdated())
}

func (f *MixinFileset) GetLastUpdated() time.Time {
	var max time.Time
	for _, f := range f.Files {
		if f.Updated.After(max) {
			max = f.Updated
		}
	}
	return max
}

type MixinFile struct {
	File    string
	URL     *url.URL
	Updated time.Time
}

// MixinEntries is used to sort the entries in a mixin feed by when they were last updated
type MixinEntries []*MixinFileset

func (e MixinEntries) Len() int {
	return len(e)
}

func (e MixinEntries) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e MixinEntries) Less(i, j int) bool {
	// Sort by LastUpdated, Mixin, Version
	entryI := e[i]
	entryJ := e[j]
	if entryI.GetLastUpdated().Equal(entryJ.GetLastUpdated()) {
		if entryI.Mixin == entryJ.Mixin {
			return entryI.Version < entryJ.Version
		}
		return entryI.Mixin < entryJ.Mixin
	}
	return entryI.GetLastUpdated().Before(entryJ.GetLastUpdated())
}
