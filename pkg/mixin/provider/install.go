package mixinprovider

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"runtime"

	"github.com/deislabs/porter/pkg/mixin"
	"github.com/pkg/errors"
)

func (p *FileSystem) Install(opts mixin.InstallOptions) (mixin.Metadata, error) {
	mixinsDir, err := p.GetMixinsDir()
	if err != nil {
		return mixin.Metadata{}, err
	}
	mixinDir := filepath.Join(mixinsDir, opts.Name)

	clientUrl := opts.GetParsedURL()
	clientUrl.Path = path.Join(clientUrl.Path, opts.Version, fmt.Sprintf("%s-%s-%s%s", opts.Name, runtime.GOOS, runtime.GOARCH, mixin.FileExt))
	clientPath := filepath.Join(mixinDir, opts.Name) + mixin.FileExt
	err = p.downloadFile(clientUrl, clientPath)
	if err != nil {
		return mixin.Metadata{}, err
	}

	runtimeUrl := opts.GetParsedURL()
	runtimeUrl.Path = path.Join(runtimeUrl.Path, opts.Version, fmt.Sprintf("%s-linux-amd64", opts.Name))
	runtimePath := filepath.Join(mixinDir, opts.Name+"-runtime")
	err = p.downloadFile(runtimeUrl, runtimePath)
	if err != nil {
		p.FileSystem.RemoveAll(mixinDir) // If the runtime download files, cleanup the mixin so it's not half installed
		return mixin.Metadata{}, err
	}

	m := mixin.Metadata{
		Name:       opts.Name,
		Dir:        mixinDir,
		ClientPath: clientPath,
	}
	return m, nil
}

func (p *FileSystem) downloadFile(url url.URL, destPath string) error {
	if p.Debug {
		fmt.Fprintf(p.Err, "Downloading %s to %s\n", url.String(), destPath)
	}

	resp, err := http.Get(url.String())
	if err != nil {
		return errors.Wrapf(err, "error downloading the mixin from %s", url.String())
	}
	if resp.StatusCode != 200 {
		return errors.Errorf("bad status returned when downloading the mixin from %s (%d)", url.String(), resp.StatusCode)
	}
	defer resp.Body.Close()

	// Ensure the parent directories exist
	parentDir := filepath.Dir(destPath)
	err = p.FileSystem.MkdirAll(parentDir, 0755)
	if err != nil {
		errors.Wrapf(err, "unable to create parent directory %s", parentDir)
	}
	cleanup := func() {
		p.FileSystem.RemoveAll(parentDir) // If we can't install the mixin, don't leave traces of it
	}

	destFile, err := p.FileSystem.Create(destPath)
	if err != nil {
		cleanup()
		return errors.Wrapf(err, "could not create the mixin at %s", destPath)
	}
	defer destFile.Close()
	err = p.FileSystem.Chmod(destPath, 0755)
	if err != nil {
		cleanup()
		return errors.Wrapf(err, "could not set the mixin as executable at %s", destPath)
	}

	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		cleanup()
		return errors.Wrapf(err, "error writing the mixin to %s", destPath)
	}
	return nil
}
