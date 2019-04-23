package mixinprovider

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg/config"
	"github.com/deislabs/porter/pkg/mixin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSystem_InstallFromUrl(t *testing.T) {
	// serve out a fake mixin
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a mixxin\n")
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	c.SetupPorterHome()
	p := NewFileSystem(c.Config)

	opts := mixin.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	m, err := p.Install(opts)

	require.NoError(t, err)
	assert.Equal(t, "mixxin", m.Name)
	assert.Equal(t, "/root/.porter/mixins/mixxin", m.Dir)
	assert.Equal(t, "/root/.porter/mixins/mixxin/mixxin", m.ClientPath)

	clientExists, _ := p.FileSystem.Exists("/root/.porter/mixins/mixxin/mixxin")
	assert.True(t, clientExists)
	runtimeExists, _ := p.FileSystem.Exists("/root/.porter/mixins/mixxin/mixxin-runtime")
	assert.True(t, runtimeExists)
}

func TestFileSystem_InstallFromFeedUrl(t *testing.T) {
	var testURL = ""
	feed, err := ioutil.ReadFile("../feed/testdata/atom.xml")
	require.NoError(t, err)

	// serve out a fake feed and mixin
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.RequestURI, "atom.xml") {
			// swap out the urls in the test atom feed to match the test http server here so that porter downloads
			// the mixin binaries from the fake server
			testAtom := strings.Replace(string(feed), "https://porter.sh", testURL, -1)
			fmt.Fprintln(w, testAtom)
		} else {
			fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am the helm mixin\n")
		}
	}))
	defer ts.Close()
	testURL = ts.URL

	c := config.NewTestConfig(t)
	c.SetupPorterHome()
	p := NewFileSystem(c.Config)

	opts := mixin.InstallOptions{
		Version: "v1.2.4",
		FeedURL: ts.URL + "/atom.xml",
	}
	opts.Validate([]string{"helm"})

	m, err := p.Install(opts)

	require.NoError(t, err)
	assert.Equal(t, "helm", m.Name)
	assert.Equal(t, "/root/.porter/mixins/helm", m.Dir)
	assert.Equal(t, "/root/.porter/mixins/helm/helm", m.ClientPath)

	clientExists, _ := p.FileSystem.Exists("/root/.porter/mixins/helm/helm")
	assert.True(t, clientExists)
	runtimeExists, _ := p.FileSystem.Exists("/root/.porter/mixins/helm/helm-runtime")
	assert.True(t, runtimeExists)
}

/*
* Revisit when we can make a general purpose new Afero FS for sabotaging
 arbitrary OS calls. This test as-is works when run as a non-root user, but
 our CI runs as root, so it's commented out.
func TestFileSystem_Install_RollbackBadDownload(t *testing.T) {
	// serve out a fake mixin
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a mixxin\n")
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config)
	// Hit the real file system for this test because Afero doesn't enforce file permissions and that's how we are
	// sabotaging the install
	c.FileSystem = &afero.Afero{Fs: afero.NewOsFs()}

	// bin is my home now
	binDir := c.TestContext.FindBinDir()
	os.Setenv(config.EnvHOME, binDir)
	defer os.Unsetenv(config.EnvHOME)

	// Make the install fail
	mixinsDir, _ := p.GetMixinsDir()
	mixinDir := path.Join(mixinsDir, "mixxin")
	p.FileSystem.MkdirAll(mixinDir, 0755)
	f, err := p.FileSystem.OpenFile(path.Join(mixinDir, "mixxin"), os.O_CREATE, 0000)
	require.NoError(t, err)
	f.Close()

	opts := mixin.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	_, err = p.Install(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not create the mixin at")

	// Make sure the mixin directory was removed
	mixinDirExists, _ := p.FileSystem.DirExists(mixinDir)
	assert.False(t, mixinDirExists)
}
*/

func TestFileSystem_Install_RollbackMissingRuntime(t *testing.T) {
	// serve out a fake mixin
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.RequestURI, "linux-amd64") {
			w.WriteHeader(400)
		} else {
			fmt.Fprintf(w, "#!/usr/bin/env bash\necho i am a client mixxin\n")
		}
	}))
	defer ts.Close()

	c := config.NewTestConfig(t)
	p := NewFileSystem(c.Config)

	mixinsDir, _ := p.GetMixinsDir()
	mixinDir := path.Join(mixinsDir, "mixxin")

	opts := mixin.InstallOptions{
		Version: "latest",
		URL:     ts.URL,
	}
	opts.Validate([]string{"mixxin"})

	_, err := p.Install(opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad status returned when downloading the mixin")

	// Make sure the mixin directory was removed
	mixinDirExists, _ := p.FileSystem.DirExists(mixinDir)
	assert.False(t, mixinDirExists)
}
