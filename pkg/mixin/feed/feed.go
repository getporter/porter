package feed

import "time"

type MixinFeed map[string]map[string]*MixinFileset

type MixinFileset struct {
	Mixin   string
	Version string
	Files   []MixinFile
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
	return e[i].GetLastUpdated().Before(e[j].GetLastUpdated())
}
