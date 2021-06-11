package releases

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"get.porter.sh/porter/mage"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/carolynvs/magex/xplat"
	"golang.org/x/sync/errgroup"
)

var (
	runtimeArch         = "amd64"
	runtimePlatform     = "linux"
	supportedClientGOOS = []string{"linux", "darwin", "windows"}
)

func getLDFLAGS(pkg string) string {
	info := mage.LoadMetadatda()
	return fmt.Sprintf("-w -X %s/pkg.Version=%s -X %s/pkg.Commit=%s", pkg, info.Version, pkg, info.Commit)
}

func build(pkg, cmd, outPath, goos, goarch string) error {
	ldflags := getLDFLAGS(pkg)

	os.MkdirAll(filepath.Dir(outPath), 0750)
	outPath += xplat.FileExt()
	srcPath := "./cmd/" + cmd

	return shx.Command("go", "build", "-ldflags", ldflags, "-o", outPath, srcPath).
		Env("CGO_ENABLED=0", "GO111MODULE=on", "GOOS="+goos, "GOARCH="+goarch).
		RunV()
}

func BuildRuntime(pkg string, name string, binDir string) error {
	outPath := filepath.Join(binDir, "runtimes", name+"-runtime")
	return build(pkg, name, outPath, runtimePlatform, runtimeArch)
}

func BuildClient(pkg string, name string, binDir string) error {
	outPath := filepath.Join(binDir, name)
	return build(pkg, name, outPath, runtime.GOOS, runtime.GOARCH)
}

func BuildAll(pkg string, name string, binDir string) error {
	var g errgroup.Group
	g.Go(func() error {
		return BuildClient(pkg, name, binDir)
	})
	g.Go(func() error {
		return BuildRuntime(pkg, name, binDir)
	})
	return g.Wait()
}

func XBuild(pkg string, name string, binDir string, goos string, goarch string) error {
	info := mage.LoadMetadatda()
	outPath := filepath.Join(binDir, info.Version, fmt.Sprintf("%s-%s-%s%s", name, goos, goarch, xplat.FileExt()))
	return build(pkg, name, outPath, goos, goarch)
}

func XBuildAll(pkg string, name string, binDir string) {
	var g errgroup.Group
	for _, goos := range supportedClientGOOS {
		goos := goos
		g.Go(func() error {
			return XBuild(pkg, name, binDir, goos, "amd64")
		})
	}

	mgx.Must(g.Wait())

	info := mage.LoadMetadatda()

	// Copy most recent build into bin/dev so that subsequent build steps can easily find it, not used for publishing
	os.RemoveAll(filepath.Join(binDir, "dev"))
	shx.Copy(filepath.Join(binDir, info.Version), filepath.Join(binDir, "dev"), shx.CopyRecursive)

	PrepareMixinForPublish(name)
}
