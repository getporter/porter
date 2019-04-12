package mixinprovider

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"

	"github.com/deislabs/porter/pkg/mixin"
)

func (p *FileSystem) Install(opts mixin.InstallOptions) error {
	mixinsDir, err := p.GetMixinsDir()
	if err != nil {
		return err
	}

	mixinDir := filepath.Join(mixinsDir, opts.Name)

	clientUrl := opts.GetParsedURL()
	clientUrl.Path = path.Join(clientUrl.Path, opts.Version, fmt.Sprintf("%s-%s-%s%s", opts.Name, runtime.GOOS, runtime.GOARCH, mixin.FileExt))
	clientPath := filepath.Join(mixinDir, opts.Name) + mixin.FileExt
	err = p.downloadFile(clientUrl, clientPath)
	if err != nil {
		return err
	}

	runtimeUrl := opts.GetParsedURL()
	runtimeUrl.Path = path.Join(runtimeUrl.Path, opts.Version, fmt.Sprintf("%s-runtime-linux-amd64", opts.Name))
	runtimePath := filepath.Join(mixinDir, opts.Name+"-runtime")
	err = p.downloadFile(runtimeUrl, runtimePath)
	if err != nil {
		return err
	}

	m := mixin.Metadata{
		Name:       opts.Name,
		Dir:        mixinDir,
		ClientPath: clientPath,
	}
	confirmedVersion, err := p.GetVersion(m)

	// TODO: Once we can extract the version from the mixin with json (#263), then we can print it out as installed mixin @v1.0.0
	if p.Debug {
		fmt.Fprintf(p.Out, "installed %s mixin to %s\n%s", m.Name, m.Dir, confirmedVersion)
	} else {
		fmt.Fprintf(p.Out, "installed %s mixin\n%s", m.Name, confirmedVersion)
	}

	return nil
}

func (p *FileSystem) downloadFile(url url.URL, destPath string) error {
	// Ensure the parent directories exist
	parentDir := filepath.Dir(destPath)
	err := os.MkdirAll(parentDir, 0755)
	if err != nil {
		errors.Wrapf(err, "unable to create parent directory %s", parentDir)
	}

	resp, err := http.Get(url.String())
	if err != nil {
		return errors.Wrapf(err, "error downloading the mixin from %s", url.String())
	}
	if resp.StatusCode != 200 {
		return errors.Errorf("bad status returned when downloading the mixin from %s (%d)", url.String(), resp.StatusCode)
	}
	defer resp.Body.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return errors.Wrapf(err, "could not create the mixin at %s", destPath)
	}
	defer destFile.Close()
	err = os.Chmod(destPath, 0755)
	if err != nil {
		return errors.Wrapf(err, "could not set the mixin as executable at %s", destPath)
	}

	if p.Debug {
		fmt.Fprintf(p.Err, "Downloading %s to %s\n", url.String(), destPath)
	}
	_, err = io.Copy(destFile, resp.Body)
	if err != nil {
		return errors.Wrapf(err, "error writing the mixin to %s", destPath)
	}
	return nil
}
