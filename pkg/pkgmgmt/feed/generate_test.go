package feed

import (
	"context"
	"fmt"
	"io/ioutil"
	"sort"
	"testing"
	"time"

	"get.porter.sh/porter/pkg/portercontext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	ctx := context.Background()
	tc := portercontext.NewTestContext(t)
	tc.AddTestFile("testdata/atom-template.xml", "template.xml")

	tc.FileSystem.Create("bin/v1.2.3/helm-darwin-amd64")
	tc.FileSystem.Create("bin/v1.2.3/helm-linux-amd64")
	tc.FileSystem.Create("bin/v1.2.3/helm-windows-amd64.exe")

	// Force the up3 timestamps to stay the same for each test run
	up3, _ := time.Parse("2006-Jan-02", "2013-Feb-03")
	tc.FileSystem.Chtimes("bin/v1.2.3/helm-darwin-amd64", up3, up3)
	tc.FileSystem.Chtimes("bin/v1.2.3/helm-linux-amd64", up3, up3)
	tc.FileSystem.Chtimes("bin/v1.2.3/helm-windows-amd64.exe", up3, up3)

	tc.FileSystem.Create("bin/v1.2.4/helm-darwin-amd64")
	tc.FileSystem.Create("bin/v1.2.4/helm-linux-amd64")
	tc.FileSystem.Create("bin/v1.2.4/helm-windows-amd64.exe")

	up4, _ := time.Parse("2006-Jan-02", "2013-Feb-04")
	tc.FileSystem.Chtimes("bin/v1.2.4/helm-darwin-amd64", up4, up4)
	tc.FileSystem.Chtimes("bin/v1.2.4/helm-linux-amd64", up4, up4)
	tc.FileSystem.Chtimes("bin/v1.2.4/helm-windows-amd64.exe", up4, up4)

	tc.FileSystem.Create("bin/v1.2.3/exec-darwin-amd64")
	tc.FileSystem.Create("bin/v1.2.3/exec-linux-amd64")
	tc.FileSystem.Create("bin/v1.2.3/exec-windows-amd64.exe")

	up2, _ := time.Parse("2006-Jan-02", "2013-Feb-02")
	tc.FileSystem.Chtimes("bin/v1.2.3/exec-darwin-amd64", up2, up2)
	tc.FileSystem.Chtimes("bin/v1.2.3/exec-linux-amd64", up2, up2)
	tc.FileSystem.Chtimes("bin/v1.2.3/exec-windows-amd64.exe", up2, up2)

	tc.FileSystem.Create("bin/canary/exec-darwin-amd64")
	tc.FileSystem.Create("bin/canary/exec-linux-amd64")
	tc.FileSystem.Create("bin/canary/exec-windows-amd64.exe")

	up10, _ := time.Parse("2006-Jan-02", "2013-Feb-10")
	tc.FileSystem.Chtimes("bin/canary/exec-darwin-amd64", up10, up10)
	tc.FileSystem.Chtimes("bin/canary/exec-linux-amd64", up10, up10)
	tc.FileSystem.Chtimes("bin/canary/exec-windows-amd64.exe", up10, up10)

	// Create extraneous release directories that should be ignored
	tc.FileSystem.Create("bin/v0.34.0-1-gda/helm-darwin-amd64")
	tc.FileSystem.Create("bin/v0.34.0-2-g1234567/helm-linux-amd64")
	tc.FileSystem.Create("bin/v0.34.0-3-g12345/helm-windows-amd64.exe")

	tc.FileSystem.Create("bin/latest/helm-darwin-amd64")
	tc.FileSystem.Create("bin/latest/helm-linux-amd64")
	tc.FileSystem.Create("bin/latest/helm-windows-amd64.exe")

	tc.FileSystem.Create("bin/canary-v1/exec-darwin-amd64")
	tc.FileSystem.Create("bin/canary-v1/exec-linux-amd64")
	tc.FileSystem.Create("bin/canary-v1/exec-windows-amd64.exe")

	opts := GenerateOptions{
		AtomFile:        "atom.xml",
		SearchDirectory: "bin",
		TemplateFile:    "template.xml",
	}
	f := NewMixinFeed(tc.Context)
	err := f.Generate(ctx, opts)
	require.NoError(t, err)
	err = f.Save(opts)
	require.NoError(t, err)

	b, err := tc.FileSystem.ReadFile("atom.xml")
	require.NoError(t, err)
	gotXml := string(b)

	b, err = ioutil.ReadFile("testdata/atom.xml")
	require.NoError(t, err)
	wantXml := string(b)

	assert.Equal(t, wantXml, gotXml)
}

func TestShouldPublishVersion(t *testing.T) {
	testcases := map[string]bool{
		"canary":                    true,
		"latest":                    false,
		"dev":                       false,
		"rando-thing":               false,
		"v1.2.3":                    true,
		"v1.2.3-alpha.1":            true,
		"v1.2.3-alpha.1+metadata":   true,
		"v1.2.3-alpha.1-1-ga368f9f": false,
		"v1.2.3-0-g":                true, // Need the hash
		"v0.33.2-0-ga":              false,
		"v0.33.2-0-g12345678":       false,
		"v0.33.2-0-ga1b3c5":         false,
	}

	for version, wantPublish := range testcases {
		t.Run(version, func(t *testing.T) {
			gotPublish := shouldPublishVersion(version)
			assert.Equal(t, wantPublish, gotPublish)
		})
	}
}

func TestGenerate_RegexMatch(t *testing.T) {
	testcases := []struct {
		name      string
		mixinName string
		wantError string
	}{{
		name:      "no bins",
		mixinName: "",
		wantError: `failed to traverse the bin directory: open /bin: file does not exist`,
	}, {
		name:      "valid mixin name",
		mixinName: "my-42nd-mixin",
		wantError: "",
	}, {
		name:      "invalid mixin name",
		mixinName: "my-42nd-mixin!",
		wantError: `no mixin binaries found in bin matching the regex "(.*/)?(.+)/([a-z0-9-]+)-(linux|windows|darwin)-(amd64)(\\.exe)?"`,
	}}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			porterCtx := portercontext.NewTestContext(t)
			porterCtx.AddTestFile("testdata/atom-template.xml", "template.xml")

			if tc.mixinName != "" {
				porterCtx.FileSystem.Create(fmt.Sprintf("bin/v1.2.3/%s-darwin-amd64", tc.mixinName))
				porterCtx.FileSystem.Create(fmt.Sprintf("bin/v1.2.3/%s-linux-amd64", tc.mixinName))
				porterCtx.FileSystem.Create(fmt.Sprintf("bin/v1.2.3/%s-windows-amd64.exe", tc.mixinName))
			}

			opts := GenerateOptions{
				AtomFile:        "atom.xml",
				SearchDirectory: "bin",
				TemplateFile:    "template.xml",
			}
			f := NewMixinFeed(porterCtx.Context)
			err := f.Generate(ctx, opts)
			if tc.wantError != "" {
				require.EqualError(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerate_ExistingFeed(t *testing.T) {
	ctx := context.Background()
	tc := portercontext.NewTestContext(t)
	tc.AddTestFile("testdata/atom-template.xml", "template.xml")
	tc.AddTestFile("testdata/atom-existing.xml", "atom.xml")

	tc.FileSystem.Create("bin/v1.2.4/helm-darwin-amd64")
	tc.FileSystem.Create("bin/v1.2.4/helm-linux-amd64")
	tc.FileSystem.Create("bin/v1.2.4/helm-windows-amd64.exe")

	up4, _ := time.Parse("2006-Jan-02", "2013-Feb-04")
	tc.FileSystem.Chtimes("bin/v1.2.4/helm-darwin-amd64", up4, up4)
	tc.FileSystem.Chtimes("bin/v1.2.4/helm-linux-amd64", up4, up4)
	tc.FileSystem.Chtimes("bin/v1.2.4/helm-windows-amd64.exe", up4, up4)

	tc.FileSystem.Create("bin/canary/exec-darwin-amd64")
	tc.FileSystem.Create("bin/canary/exec-linux-amd64")
	tc.FileSystem.Create("bin/canary/exec-windows-amd64.exe")

	up10, _ := time.Parse("2006-Jan-02", "2013-Feb-10")
	tc.FileSystem.Chtimes("bin/canary/exec-darwin-amd64", up10, up10)
	tc.FileSystem.Chtimes("bin/canary/exec-linux-amd64", up10, up10)
	tc.FileSystem.Chtimes("bin/canary/exec-windows-amd64.exe", up10, up10)

	opts := GenerateOptions{
		AtomFile:        "atom.xml",
		SearchDirectory: "bin",
		TemplateFile:    "template.xml",
	}
	f := NewMixinFeed(tc.Context)
	err := f.Generate(ctx, opts)
	require.NoError(t, err)
	err = f.Save(opts)
	require.NoError(t, err)

	b, err := tc.FileSystem.ReadFile("atom.xml")
	require.NoError(t, err)
	gotXml := string(b)

	b, err = ioutil.ReadFile("testdata/atom.xml")
	require.NoError(t, err)
	wantXml := string(b)

	assert.Equal(t, wantXml, gotXml)
}

func TestGenerate_RegenerateDoesNotCreateDuplicates(t *testing.T) {
	ctx := context.Background()
	tc := portercontext.NewTestContext(t)
	tc.AddTestFile("testdata/atom-template.xml", "template.xml")
	tc.AddTestFile("testdata/atom-existing.xml", "atom.xml")

	tc.FileSystem.Create("bin/v1.2.4/helm-darwin-amd64")
	tc.FileSystem.Create("bin/v1.2.4/helm-linux-amd64")
	tc.FileSystem.Create("bin/v1.2.4/helm-windows-amd64.exe")

	up4, _ := time.Parse("2006-Jan-02", "2013-Feb-04")
	tc.FileSystem.Chtimes("bin/v1.2.4/helm-darwin-amd64", up4, up4)
	tc.FileSystem.Chtimes("bin/v1.2.4/helm-linux-amd64", up4, up4)
	tc.FileSystem.Chtimes("bin/v1.2.4/helm-windows-amd64.exe", up4, up4)

	tc.FileSystem.Create("bin/canary/exec-darwin-amd64")
	tc.FileSystem.Create("bin/canary/exec-linux-amd64")
	tc.FileSystem.Create("bin/canary/exec-windows-amd64.exe")

	up10, _ := time.Parse("2006-Jan-02", "2013-Feb-10")
	tc.FileSystem.Chtimes("bin/canary/exec-darwin-amd64", up10, up10)
	tc.FileSystem.Chtimes("bin/canary/exec-linux-amd64", up10, up10)
	tc.FileSystem.Chtimes("bin/canary/exec-windows-amd64.exe", up10, up10)

	opts := GenerateOptions{
		AtomFile:        "atom.xml",
		SearchDirectory: "bin",
		TemplateFile:    "template.xml",
	}
	f := NewMixinFeed(tc.Context)

	err := f.Generate(ctx, opts)
	require.NoError(t, err)
	err = f.Save(opts)
	require.NoError(t, err)

	// Run the generation again, against the same versions, and make sure they don't insert duplicate files
	// This mimics what the CI does when we repeat a build, or have multiple
	// canary builds on the "main" branch
	err = f.Generate(ctx, opts)
	require.NoError(t, err)
	err = f.Save(opts)
	require.NoError(t, err)

	b, err := tc.FileSystem.ReadFile("atom.xml")
	require.NoError(t, err)
	gotXml := string(b)

	b, err = ioutil.ReadFile("testdata/atom.xml")
	require.NoError(t, err)
	wantXml := string(b)

	assert.Equal(t, wantXml, gotXml)
}

func TestMixinEntries_Sort(t *testing.T) {
	up2, _ := time.Parse("2006-Jan-02", "2013-Feb-02")
	up3, _ := time.Parse("2006-Jan-02", "2013-Feb-03")
	up4, _ := time.Parse("2006-Jan-02", "2013-Feb-04")

	entries := MixinEntries{
		{
			Files: []*MixinFile{
				{Updated: up3},
			},
		},
		{
			Files: []*MixinFile{
				{Updated: up2},
			},
		},
		{
			Files: []*MixinFile{
				{Updated: up4},
			},
		},
	}

	sort.Sort(sort.Reverse(entries))

	assert.Equal(t, up4, entries[0].Files[0].Updated)
	assert.Equal(t, up3, entries[1].Files[0].Updated)
	assert.Equal(t, up2, entries[2].Files[0].Updated)
}
