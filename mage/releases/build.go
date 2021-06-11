package releases

import (
	"fmt"
	"os"
	"path/filepath"

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

func BuildRuntime(pkg string, name string, binDir string) error {
	ldflags := getLDFLAGS(pkg)

	runtimeDir := filepath.Join(binDir, "runtimes")
	os.MkdirAll(runtimeDir, 0750)
	return shx.Command("go", "build", "-ldflags", ldflags, "-o", filepath.Join(runtimeDir, name+"-runtime"+xplat.FileExt()), "./cmd/"+name).
		Env("GO111MODULE=on", "GOARCH="+runtimeArch, "GOOS="+runtimePlatform, "CGO_ENABLED=0").
		RunV()
}

func BuildClient(pkg string, name string, binDir string) error {
	ldflags := getLDFLAGS(pkg)

	os.MkdirAll(binDir, 0750)
	return shx.Command("go", "build", "-ldflags", ldflags, "-o", filepath.Join(binDir, name+xplat.FileExt()), "./cmd/"+name).
		Env("GO111MODULE=on", "CGO_ENABLED=0").
		RunV()
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
	ldflags := getLDFLAGS(pkg)
	info := mage.LoadMetadatda()
	outputPath := filepath.Join(binDir, info.Version, fmt.Sprintf("%s-%s-%s%s", name, goos, goarch, xplat.FileExt()))
	os.MkdirAll(filepath.Dir(outputPath), 0750)
	return shx.Command("go", "build", "-ldflags", ldflags, "-o", outputPath, "./cmd/"+name).
		Env("GO111MODULE=on", "GOARCH="+goarch, "GOOS="+goos, "CGO_ENABLED=0").
		RunV()
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
