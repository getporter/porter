package mixinprovider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"runtime"

	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/mixin/feed"
	"github.com/pkg/errors"
)

func (fs *FileSystem) Install(opts mixin.InstallOptions) (*mixin.Metadata, error) {
	var err error
	var metadata *mixin.Metadata
	if opts.FeedURL != "" {
		metadata, err = fs.InstallFromFeedURL(opts)
	} else {
		metadata, err = fs.InstallFromURL(opts)
	}
	if err != nil {
		return nil, err
	}
	err = fs.saveMixinInfo(opts)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

const MixinCacheJSON string = "cache.json"

func (fs *FileSystem) saveMixinInfo(opts mixin.InstallOptions) error {
	mixinsDir, _ := fs.Config.GetMixinsDir()
	cacheJSONPath := filepath.Join(mixinsDir, "/", MixinCacheJSON)
	exists, _ := fs.FileSystem.Exists(cacheJSONPath)
	if !exists {
		_, err := fs.FileSystem.Create(cacheJSONPath)
		if err != nil {
			return errors.Wrap(err, "error creating mixin cache.json")
		}
	}

	mixinCacheFileContents, err := fs.FileSystem.ReadFile(cacheJSONPath)
	if err != nil {
		return errors.Wrap(err, "error reading mixin cache.json")
	}

	mixinsDataJSON := &mixins{}
	if len(mixinCacheFileContents) > 0 {
		err = json.Unmarshal(mixinCacheFileContents, &mixinsDataJSON)
		if err != nil {
			return errors.Wrap(err, "error unmarshalling from mixin cache.json")
		}
	}
	//if a mixin exists, skip.
	for _, mixin := range mixinsDataJSON.Mixins {
		if mixin.Name == opts.Name {
			return nil
		}
	}
	updatedMixinList := append(mixinsDataJSON.Mixins, mixinInfo{Name: opts.Name, FeedURL: opts.FeedURL, URL: opts.URL})
	mixinsDataJSON.Mixins = updatedMixinList
	updatedMixinInfo, err := json.MarshalIndent(&mixinsDataJSON, "", "  ")
	if err != nil {
		return errors.Wrap(err, "error marshalling to mixin cache.json")
	}
	err = fs.FileSystem.WriteFile(cacheJSONPath, updatedMixinInfo, 0644)

	if err != nil {
		return errors.Wrap(err, "error adding mixin info cache.json")
	}
	return nil
}

type mixinInfo struct {
	Name    string `json:"name"`
	FeedURL string `json:"feedURL,omitempty"`
	URL     string `json:"url,omitempty"`
}

type mixins struct {
	Mixins []mixinInfo `json:"mixins"`
}

func (fs *FileSystem) InstallFromURL(opts mixin.InstallOptions) (*mixin.Metadata, error) {
	clientUrl := opts.GetParsedURL()
	clientUrl.Path = path.Join(clientUrl.Path, opts.Version, fmt.Sprintf("%s-%s-%s%s", opts.Name, runtime.GOOS, runtime.GOARCH, mixin.FileExt))

	runtimeUrl := opts.GetParsedURL()
	runtimeUrl.Path = path.Join(runtimeUrl.Path, opts.Version, fmt.Sprintf("%s-linux-amd64", opts.Name))

	return fs.downloadMixin(opts.Name, clientUrl, runtimeUrl)
}

func (fs *FileSystem) InstallFromFeedURL(opts mixin.InstallOptions) (*mixin.Metadata, error) {
	feedUrl := opts.GetParsedFeedURL()
	tmpDir, err := fs.FileSystem.TempDir("", "porter")
	if err != nil {
		return nil, errors.Wrap(err, "error creating temp directory")
	}
	defer fs.FileSystem.RemoveAll(tmpDir)
	feedPath := filepath.Join(tmpDir, "atom.xml")

	err = fs.downloadFile(feedUrl, feedPath, false)
	if err != nil {
		return nil, err
	}

	searchFeed := feed.NewMixinFeed(fs.Context)
	err = searchFeed.Load(feedPath)
	if err != nil {
		return nil, err
	}

	result := searchFeed.Search(opts.Name, opts.Version)
	if result == nil {
		return nil, errors.Errorf("the mixin feed at %s does not contain an entry for %s @ %s", opts.FeedURL, opts.Name, opts.Version)
	}

	clientUrl := result.FindDownloadURL(runtime.GOOS, runtime.GOARCH)
	if clientUrl == nil {
		return nil, errors.Errorf("%s @ %s did not publish a download for %s/%s", opts.Name, opts.Version, runtime.GOOS, runtime.GOARCH)
	}

	runtimeUrl := result.FindDownloadURL("linux", "amd64")
	if runtimeUrl == nil {
		return nil, errors.Errorf("%s @ %s did not publish a download for linux/amd64", opts.Name, opts.Version)
	}

	return fs.downloadMixin(opts.Name, *clientUrl, *runtimeUrl)
}

func (fs *FileSystem) downloadMixin(name string, clientUrl url.URL, runtimeUrl url.URL) (*mixin.Metadata, error) {
	mixinsDir, err := fs.GetMixinsDir()
	if err != nil {
		return nil, err
	}
	mixinDir := filepath.Join(mixinsDir, name)

	clientPath := filepath.Join(mixinDir, name) + mixin.FileExt
	err = fs.downloadFile(clientUrl, clientPath, true)
	if err != nil {
		return nil, err
	}

	runtimePath := filepath.Join(mixinDir, name+"-runtime")
	err = fs.downloadFile(runtimeUrl, runtimePath, true)
	if err != nil {
		fs.FileSystem.RemoveAll(mixinDir) // If the runtime download fails, cleanup the mixin so it's not half installed
		return nil, err
	}

	m := mixin.Metadata{
		Name:       name,
		Dir:        mixinDir,
		ClientPath: clientPath,
	}
	return &m, nil
}

func (fs *FileSystem) downloadFile(url url.URL, destPath string, executable bool) error {
	if fs.Debug {
		fmt.Fprintf(fs.Err, "Downloading %s to %s\n", url.String(), destPath)
	}

	resp, err := http.Get(url.String())
	if err != nil {
		return errors.Wrapf(err, "error downloading %s", url.String())
	}
	if resp.StatusCode != 200 {
		return errors.Errorf("bad status returned when downloading %s (%d)", url.String(), resp.StatusCode)
	}
	defer resp.Body.Close()

	// Ensure the parent directories exist
	parentDir := filepath.Dir(destPath)
	parentDirExists, err := fs.FileSystem.DirExists(parentDir)
	if err != nil {
		return errors.Wrapf(err, "unable to check if directory exists %s", parentDir)
	}

	cleanup := func() {}
	if !parentDirExists {
		err = fs.FileSystem.MkdirAll(parentDir, 0755)
		if err != nil {
			errors.Wrapf(err, "unable to create parent directory %s", parentDir)
		}
		cleanup = func() {
			fs.FileSystem.RemoveAll(parentDir) // If we can't download the file, don't leave traces of it
		}
	}

	destFile, err := fs.FileSystem.Create(destPath)
	if err != nil {
		cleanup()
		return errors.Wrapf(err, "could not create the file at %s", destPath)
	}
	defer destFile.Close()

	if executable {
		err = fs.FileSystem.Chmod(destPath, 0755)
		if err != nil {
			cleanup()
			return errors.Wrapf(err, "could not set the file as executable at %s", destPath)
		}
	}

	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		cleanup()
		return errors.Wrapf(err, "error writing the file to %s", destPath)
	}
	return nil
}
