package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_InstallFromUrl(t *testing.T) {
	// serve out a fake package
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Azure-DebugInfo") == "" ||
			!strings.Contains(r.Header.Get("User-Agent"), "porter_trace") {
			w.WriteHeader(400)
		}
		fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a mixxin\n")
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config, "packages")

	opts := pkgmgmt.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	err := p.Install(opts)
	require.NoError(t, err)

	clientExists, _ := p.FileSystem.Exists("/root/.porter/packages/mixxin/mixxin")
	assert.True(t, clientExists)
	runtimeExists, _ := p.FileSystem.Exists("/root/.porter/packages/mixxin/runtimes/mixxin-runtime")
	assert.True(t, runtimeExists)
}

func TestFileSystem_InstallFromFeedUrl(t *testing.T) {
	var testURL = ""
	feed, err := ioutil.ReadFile("../feed/testdata/atom.xml")
	require.NoError(t, err)

	// serve out a fake feed and package
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.RequestURI, "atom.xml") {
			// swap out the urls in the test atom feed to match the test http server here so that porter downloads
			// the package binaries from the fake server
			testAtom := strings.Replace(string(feed), "https://cdn.porter.sh", testURL, -1)
			fmt.Fprintln(w, testAtom)
		} else {
			fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am helm\n")
		}
	}))
	defer ts.Close()
	testURL = ts.URL

	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config, "packages")

	opts := pkgmgmt.InstallOptions{
		Version: "v1.2.4",
		FeedURL: ts.URL + "/atom.xml",
	}
	opts.Validate([]string{"helm"})

	err = p.Install(opts)
	require.NoError(t, err)

	clientExists, _ := p.FileSystem.Exists("/root/.porter/packages/helm/helm")
	assert.True(t, clientExists)
	runtimeExists, _ := p.FileSystem.Exists("/root/.porter/packages/helm/runtimes/helm-runtime")
	assert.True(t, runtimeExists)
}

func TestFileSystem_Install_RollbackMissingRuntime(t *testing.T) {
	// serve out a fake package
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, "linux-amd64") {
			w.WriteHeader(400)
		} else {
			fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a client mixxin\n")
		}
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config, "packages")

	parentDir, _ := p.GetPackagesDir()
	pkgDir := path.Join(parentDir, "mixxin")

	opts := pkgmgmt.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	err := p.Install(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad status returned when downloading")
	assert.Contains(t, err.Error(), "porter_trace_", "The error message should contain our special debug user agent")

	// Make sure the package directory was removed
	dirExists, _ := p.FileSystem.DirExists(pkgDir)
	assert.False(t, dirExists)
}

func TestFileSystem_Install_PackageInfoSavedWhenNoFileExists(t *testing.T) {
	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config, "packages")

	packageURL := "https://cdn.porter.sh/mixins/helm"
	opts := pkgmgmt.InstallOptions{
		Version: "v1.2.4",
		URL:     packageURL,
	}
	name := "helm"
	opts.Validate([]string{name})

	// ensure cache.json does not exist (yet)
	cacheExists, _ := p.FileSystem.Exists("/root/.porter/packages/cache.json")
	assert.False(t, cacheExists)

	err := p.savePackageInfo(opts)
	require.NoError(t, err)

	// cache.json should have been created
	cacheExists, _ = p.FileSystem.Exists("/root/.porter/packages/cache.json")
	assert.True(t, cacheExists)

	cacheContentsB, err := p.FileSystem.ReadFile("/root/.porter/packages/cache.json")
	require.NoError(t, err)

	//read cache.json
	var allPackages packages
	err = json.Unmarshal(cacheContentsB, &allPackages)
	require.NoError(t, err)

	//confirm that the required pkg is present
	var pkgData PackageInfo
	for _, pkg := range allPackages.Packages {
		if pkg.Name == name {
			pkgData = pkg
			break
		}
	}

	assert.Equal(t, name, pkgData.Name)
	assert.Equal(t, packageURL, pkgData.URL)
}
