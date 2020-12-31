package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"runtime"

	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/feed"
	"github.com/pkg/errors"
)

const PackageCacheJSON string = "cache.json"

func (fs *FileSystem) Install(opts pkgmgmt.InstallOptions) error {
	var err error
	if opts.FeedURL != "" {
		err = fs.InstallFromFeedURL(opts)
	} else {
		err = fs.InstallFromURL(opts)
	}
	if err != nil {
		return err
	}
	return fs.savePackageInfo(opts)
}

func (fs *FileSystem) savePackageInfo(opts pkgmgmt.InstallOptions) error {
	cacheJSONPath := filepath.Join(fs.GetPackagesDir(), "/", PackageCacheJSON)
	exists, _ := fs.FileSystem.Exists(cacheJSONPath)
	if !exists {
		_, err := fs.FileSystem.Create(cacheJSONPath)
		if err != nil {
			return errors.Wrapf(err, "error creating %s package cache.json", fs.PackageType)
		}
	}

	cacheContentsB, err := fs.FileSystem.ReadFile(cacheJSONPath)
	if err != nil {
		return errors.Wrapf(err, "error reading package %s cache.json", fs.PackageType)
	}

	pkgDataJSON := &packages{}
	if len(cacheContentsB) > 0 {
		err = json.Unmarshal(cacheContentsB, &pkgDataJSON)
		if err != nil {
			return errors.Wrapf(err, "error unmarshalling from %s package cache.json", fs.PackageType)
		}
	}
	//if a package exists, skip.
	for _, pkg := range pkgDataJSON.Packages {
		if pkg.Name == opts.Name {
			return nil
		}
	}
	updatedPkgList := append(pkgDataJSON.Packages, PackageInfo{Name: opts.Name, FeedURL: opts.FeedURL, URL: opts.URL})
	pkgDataJSON.Packages = updatedPkgList
	updatedPkgInfo, err := json.MarshalIndent(&pkgDataJSON, "", "  ")
	if err != nil {
		return errors.Wrapf(err, "error marshalling to %s package cache.json", fs.PackageType)
	}
	err = fs.FileSystem.WriteFile(cacheJSONPath, updatedPkgInfo, 0644)

	if err != nil {
		return errors.Wrapf(err, "error adding package info to %s cache.json", fs.PackageType)
	}
	return nil
}

type PackageInfo struct {
	Name    string `json:"name"`
	FeedURL string `json:"URL,omitempty"`
	URL     string `json:"url,omitempty"`
}

type packages struct {
	Packages []PackageInfo `json:"packages"`
}

func (fs *FileSystem) InstallFromURL(opts pkgmgmt.InstallOptions) error {
	clientUrl := opts.GetParsedURL()
	clientUrl.Path = path.Join(clientUrl.Path, opts.Version, fmt.Sprintf("%s-%s-%s%s", opts.Name, runtime.GOOS, runtime.GOARCH, pkgmgmt.FileExt))

	runtimeUrl := opts.GetParsedURL()
	runtimeUrl.Path = path.Join(runtimeUrl.Path, opts.Version, fmt.Sprintf("%s-linux-amd64", opts.Name))

	return fs.downloadPackage(opts.Name, clientUrl, runtimeUrl)
}

func (fs *FileSystem) InstallFromFeedURL(opts pkgmgmt.InstallOptions) error {
	feedUrl := opts.GetParsedFeedURL()
	tmpDir, err := fs.FileSystem.TempDir("", "porter")
	if err != nil {
		return errors.Wrap(err, "error creating temp directory")
	}
	defer fs.FileSystem.RemoveAll(tmpDir)
	feedPath := filepath.Join(tmpDir, "atom.xml")

	err = fs.downloadFile(feedUrl, feedPath, false)
	if err != nil {
		return err
	}

	searchFeed := feed.NewMixinFeed(fs.Context)
	err = searchFeed.Load(feedPath)
	if err != nil {
		return err
	}

	result := searchFeed.Search(opts.Name, opts.Version)
	if result == nil {
		return errors.Errorf("the feed at %s does not contain an entry for %s @ %s", opts.FeedURL, opts.Name, opts.Version)
	}

	clientUrl := result.FindDownloadURL(runtime.GOOS, runtime.GOARCH)
	if clientUrl == nil {
		return errors.Errorf("%s @ %s did not publish a download for %s/%s", opts.Name, opts.Version, runtime.GOOS, runtime.GOARCH)
	}

	runtimeUrl := result.FindDownloadURL("linux", "amd64")
	if runtimeUrl == nil {
		return errors.Errorf("%s @ %s did not publish a download for linux/amd64", opts.Name, opts.Version)
	}

	return fs.downloadPackage(opts.Name, *clientUrl, *runtimeUrl)
}

func (fs *FileSystem) downloadPackage(name string, clientUrl url.URL, runtimeUrl url.URL) error {
	pkgDir := filepath.Join(fs.GetPackagesDir(), name)

	clientPath := fs.BuildClientPath(pkgDir, name)
	err := fs.downloadFile(clientUrl, clientPath, true)
	if err != nil {
		return err
	}

	runtimePath := filepath.Join(pkgDir, "runtimes", name+"-runtime")
	err = fs.downloadFile(runtimeUrl, runtimePath, true)
	if err != nil {
		fs.FileSystem.RemoveAll(pkgDir) // If the runtime download fails, cleanup the package so it's not half installed
		return err
	}

	return nil
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
