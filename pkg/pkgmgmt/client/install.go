package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"runtime"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/feed"
	"get.porter.sh/porter/pkg/tracing"
)

const PackageCacheJSON string = "cache.json"

func (fs *FileSystem) Install(ctx context.Context, opts pkgmgmt.InstallOptions) error {
	var err error
	if opts.FeedURL != "" {
		err = fs.InstallFromFeedURL(ctx, opts)
	} else {
		err = fs.InstallFromURL(ctx, opts)
	}
	if err != nil {
		return err
	}
	return fs.savePackageInfo(ctx, opts)
}

func (fs *FileSystem) savePackageInfo(ctx context.Context, opts pkgmgmt.InstallOptions) error {
	log := tracing.LoggerFromContext(ctx)

	parentDir, _ := fs.GetPackagesDir()
	cacheJSONPath := filepath.Join(parentDir, "/", PackageCacheJSON)
	exists, _ := fs.FileSystem.Exists(cacheJSONPath)
	if !exists {
		_, err := fs.FileSystem.Create(cacheJSONPath)
		if err != nil {
			return log.Error(fmt.Errorf("error creating %s package cache.json: %w", fs.PackageType, err))
		}
	}

	cacheContentsB, err := fs.FileSystem.ReadFile(cacheJSONPath)
	if err != nil {
		return log.Error(fmt.Errorf("error reading package %s cache.json: %w", fs.PackageType, err))
	}

	pkgDataJSON := &packages{}
	if len(cacheContentsB) > 0 {
		err = json.Unmarshal(cacheContentsB, &pkgDataJSON)
		if err != nil {
			return log.Error(fmt.Errorf("error unmarshalling from %s package cache.json: %w", fs.PackageType, err))
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
		return log.Error(fmt.Errorf("error marshalling to %s package cache.json: %w", fs.PackageType, err))
	}
	err = fs.FileSystem.WriteFile(cacheJSONPath, updatedPkgInfo, pkg.FileModeWritable)

	if err != nil {
		return log.Error(fmt.Errorf("error adding package info to %s cache.json: %w", fs.PackageType, err))
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

func (fs *FileSystem) InstallFromURL(ctx context.Context, opts pkgmgmt.InstallOptions) error {
	return fs.installFromURLFor(ctx, opts, runtime.GOOS, runtime.GOARCH)
}

func (fs *FileSystem) installFromURLFor(ctx context.Context, opts pkgmgmt.InstallOptions, os string, arch string) error {
	log := tracing.LoggerFromContext(ctx)

	clientUrl := opts.GetParsedURL()
	clientUrl.Path = path.Join(clientUrl.Path, opts.Version, fmt.Sprintf("%s-%s-%s%s", opts.Name, os, arch, pkgmgmt.FileExt))

	runtimeUrl := opts.GetParsedURL()
	runtimeUrl.Path = path.Join(runtimeUrl.Path, opts.Version, fmt.Sprintf("%s-linux-amd64", opts.Name))

	err := fs.downloadPackage(ctx, opts.Name, clientUrl, runtimeUrl)
	if err != nil && os == "darwin" && arch == "arm64" {
		// Until we have full support for M1 chipsets, rely on rossetta functionality in macos and use the amd64 binary
		log.Debugf("%s @ %s did not publish a download for darwin/amd64, falling back to darwin/amd64", opts.Name, opts.Version)
		return fs.installFromURLFor(ctx, opts, "darwin", "amd64")
	}

	return err
}

func (fs *FileSystem) InstallFromFeedURL(ctx context.Context, opts pkgmgmt.InstallOptions) error {
	log := tracing.LoggerFromContext(ctx)

	feedUrl := opts.GetParsedFeedURL()
	tmpDir, err := fs.FileSystem.TempDir("", "porter")
	if err != nil {
		return log.Error(fmt.Errorf("error creating temp directory: %w", err))
	}
	defer fs.FileSystem.RemoveAll(tmpDir)
	feedPath := filepath.Join(tmpDir, "atom.xml")

	err = fs.downloadFile(ctx, feedUrl, feedPath, false)
	if err != nil {
		return err
	}

	searchFeed := feed.NewMixinFeed(fs.Context)
	err = searchFeed.Load(ctx, feedPath)
	if err != nil {
		return err
	}

	result := searchFeed.Search(opts.Name, opts.Version)
	if result == nil {
		return log.Error(fmt.Errorf("the feed at %s does not contain an entry for %s @ %s", opts.FeedURL, opts.Name, opts.Version))
	}

	clientUrl := result.FindDownloadURL(ctx, runtime.GOOS, runtime.GOARCH)
	if clientUrl == nil {
		return log.Error(fmt.Errorf("%s @ %s did not publish a download for %s/%s", opts.Name, opts.Version, runtime.GOOS, runtime.GOARCH))
	}

	runtimeUrl := result.FindDownloadURL(ctx, "linux", "amd64")
	if runtimeUrl == nil {
		return log.Error(fmt.Errorf("%s @ %s did not publish a download for linux/amd64", opts.Name, opts.Version))
	}

	return fs.downloadPackage(ctx, opts.Name, *clientUrl, *runtimeUrl)
}

func (fs *FileSystem) downloadPackage(ctx context.Context, name string, clientUrl url.URL, runtimeUrl url.URL) error {
	parentDir, err := fs.GetPackagesDir()
	if err != nil {
		return err
	}
	pkgDir := filepath.Join(parentDir, name)

	clientPath := fs.BuildClientPath(pkgDir, name)
	err = fs.downloadFile(ctx, clientUrl, clientPath, true)
	if err != nil {
		return err
	}

	runtimePath := filepath.Join(pkgDir, "runtimes", name+"-runtime")
	err = fs.downloadFile(ctx, runtimeUrl, runtimePath, true)
	if err != nil {
		fs.FileSystem.RemoveAll(pkgDir) // If the runtime download fails, cleanup the package so it's not half installed
		return err
	}

	return nil
}

func (fs *FileSystem) downloadFile(ctx context.Context, url url.URL, destPath string, executable bool) error {
	log := tracing.LoggerFromContext(ctx)
	log.Debugf("Downloading %s to %s\n", url.String(), destPath)

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return log.Error(fmt.Errorf("error creating web request to %s: %w", url.String(), err))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return log.Error(fmt.Errorf("error downloading %s: %w", url.String(), err))
	}
	if resp.StatusCode != 200 {
		err := fmt.Errorf("bad status returned when downloading %s (%d) %s", url.String(), resp.StatusCode, resp.Status)
		log.Debugf(err.Error()) // Only debug log this since higher up on the stack we may handle this error
		return err
	}
	defer resp.Body.Close()

	// Ensure the parent directories exist
	parentDir := filepath.Dir(destPath)
	parentDirExists, err := fs.FileSystem.DirExists(parentDir)
	if err != nil {
		return log.Error(fmt.Errorf("unable to check if directory exists %s: %w", parentDir, err))
	}

	cleanup := func() {}
	if !parentDirExists {
		err = fs.FileSystem.MkdirAll(parentDir, pkg.FileModeDirectory)
		if err != nil {
			return log.Error(fmt.Errorf("unable to create parent directory %s: %w", parentDir, err))
		}
		cleanup = func() {
			fs.FileSystem.RemoveAll(parentDir) // If we can't download the file, don't leave traces of it
		}
	}

	destFile, err := fs.FileSystem.Create(destPath)
	if err != nil {
		cleanup()
		return log.Error(fmt.Errorf("could not create the file at %s: %w", destPath, err))
	}
	defer destFile.Close()

	if executable {
		err = fs.FileSystem.Chmod(destPath, pkg.FileModeExecutable)
		if err != nil {
			cleanup()
			return log.Error(fmt.Errorf("could not set the file as executable at %s: %w", destPath, err))
		}
	}

	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		cleanup()
		return log.Error(fmt.Errorf("error writing the file to %s: %w", destPath, err))
	}
	return nil
}
