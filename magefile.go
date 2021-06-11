// +build mage

// This is a magefile, and is a "makefile for go".
// See https://magefile.org/
package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	// mage:import
	"get.porter.sh/porter/mage/releases"

	"get.porter.sh/porter/mage"
	"get.porter.sh/porter/mage/tests"
	"get.porter.sh/porter/mage/tools"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/pkg/gopath"
	"github.com/carolynvs/magex/shx"
	"github.com/carolynvs/magex/xplat"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
// var Default = Build

const (
	mixinsURL = "https://cdn.porter.sh/mixins/"
)

var must = shx.CommandBuilder{StopOnError: true}

// Cleanup workspace after building or running tests.
func Clean() {
	mg.Deps(tests.DeleteTestCluster)
	mgx.Must(os.RemoveAll("bin"))
}

// Ensure EnsureMage is installed and on the PATH.
func EnsureMage() error {
	return tools.EnsureMage()
}

func Debug() {
	mage.LoadMetadatda()
}

// ConfigureAgent sets up an Azure DevOps agent with EnsureMage and ensures
// that GOPATH/bin is in PATH.
func ConfigureAgent() error {
	err := EnsureMage()
	if err != nil {
		return err
	}

	// Instruct Azure DevOps to add GOPATH/bin to PATH
	gobin := gopath.GetGopathBin()
	err = os.MkdirAll(gobin, 0755)
	if err != nil {
		return errors.Wrapf(err, "could not mkdir -p %s", gobin)
	}
	fmt.Printf("Adding %s to the PATH\n", gobin)
	fmt.Printf("##vso[task.prependpath]%s\n", gobin)
	return nil
}

// Install mixins used by tests and example bundles, if not already installed
func GetMixins() error {
	mixinTag := os.Getenv("MIXIN_TAG")
	if mixinTag == "" {
		mixinTag = "canary"
	}

	mixins := []string{"helm", "arm", "terraform", "kubernetes"}
	var errG errgroup.Group
	for _, mixin := range mixins {
		mixinDir := filepath.Join("bin/mixins/", mixin)
		if _, err := os.Stat(mixinDir); err == nil {
			log.Println("Mixin already installed into bin:", mixin)
			continue
		}

		mixin := mixin
		errG.Go(func() error {
			log.Println("Installing mixin:", mixin)
			mixinURL := mixinsURL + mixin
			return porter("mixin", "install", mixin, "--version", mixinTag, "--url", mixinURL).Run()
		})
	}

	return errG.Wait()
}

// Run a porter command from the bin
func porter(args ...string) shx.PreparedCommand {
	porterPath := filepath.Join("bin", "porter")
	p := shx.Command(porterPath, args...)

	porterHome, _ := filepath.Abs("bin")
	p.Cmd.Env = []string{"PORTER_HOME=" + porterHome}

	return p
}

// Update golden test files to match the new test outputs
func UpdateTestfiles() {
	must.Command("make", "test-unit").Env("PORTER_UPDATE_TEST_FILES=true").RunV()
	must.Command("make", "test-unit").RunV()
}

// Run smoke tests to quickly check if Porter is broken
func TestSmoke() error {
	mg.Deps(tests.StartDockerRegistry)
	defer tests.StopDockerRegistry()

	// Only do verbose output of tests when called with `mage -v TestSmoke`
	v := ""
	if mg.Verbose() {
		v = "-v"
	}

	return shx.Command("go", "test", "-tags", "smoke", v, "./tests/smoke/...").CollapseArgs().RunV()
}

func getRegistry() string {
	registry := os.Getenv("REGISTRY")
	if registry == "" {
		registry = "getporterci"
	}
	return registry
}

func getDualPublish() bool {
	dualPublish, _ := strconv.ParseBool(os.Getenv("DUAL_PUBLISH"))
	return dualPublish
}

func BuildImages() {
	info := mage.LoadMetadatda()

	must.Command("./scripts/build-images.sh").Env("VERSION="+info.Version, "PERMALINK="+info.Permalink, "REGISTRY="+getRegistry()).RunV()
	if getDualPublish() {
		must.Command("./scripts/build-images.sh").Env("VERSION="+info.Version, "PERMALINK="+info.Permalink, "REGISTRY=ghcr.io/getporter").RunV()
	}
}

func PublishImages() {
	mg.Deps(BuildImages)

	info := mage.LoadMetadatda()

	must.Command("./scripts/publish-images.sh").Env("VERSION="+info.Version, "PERMALINK="+info.Permalink, "REGISTRY="+getRegistry()).RunV()
	if getDualPublish() {
		must.Command("./scripts/publish-images.sh").Env("VERSION="+info.Version, "PERMALINK="+info.Permalink, "REGISTRY=ghcr.io/getporter").RunV()
	}
}

// Publish the porter binaries and install scripts.
func PublishPorter() {
	mg.Deps(tools.EnsureGitHubClient, releases.ConfigureGitBot)

	info := mage.LoadMetadatda()

	// Copy install scripts into version directory
	must.Command("./scripts/prep-install-scripts.sh").Env("VERSION="+info.Version, "PERMALINK="+info.Permalink).RunV()

	porterVersionDir := filepath.Join("bin", info.Version)
	execVersionDir := filepath.Join("bin/mixins/exec", info.Version)
	var repo = os.Getenv("PORTER_RELEASE_REPOSITORY")
	if repo == "" {
		repo = "github.com/getporter/porter"
	}
	remote := fmt.Sprintf("https://%s.git", repo)

	// Move the permalink tag. The existing release automatically points to the tag.
	must.RunV("git", "tag", info.Permalink, info.Version+"^{}", "-f")
	must.RunV("git", "push", "-f", remote, info.Permalink)

	// Create or update GitHub release for the permalink (canary/latest) with the version's assets (porter binaries, exec binaries and install scripts)
	releases.AddFilesToRelease(repo, info.Permalink, porterVersionDir)
	releases.AddFilesToRelease(repo, info.Permalink, execVersionDir)

	if info.IsTaggedRelease {
		// Create GitHub release for the exact version (v1.2.3) and attach assets
		releases.AddFilesToRelease(repo, info.Version, porterVersionDir)
		releases.AddFilesToRelease(repo, info.Version, execVersionDir)
	}
}

// Copy the cross-compiled binaries from xbuild into bin.
func UseXBuildBinaries() error {
	pwd, _ := os.Getwd()
	goos := build.Default.GOOS
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	copies := map[string]string{
		"bin/dev/porter-$GOOS-amd64$EXT":           "bin/porter$EXT",
		"bin/dev/porter-linux-amd64":               "bin/runtimes/porter-runtime",
		"bin/mixins/exec/dev/exec-$GOOS-amd64$EXT": "bin/mixins/exec/exec$EXT",
		"bin/mixins/exec/dev/exec-linux-amd64":     "bin/mixins/exec/runtimes/exec-runtime",
	}

	r := strings.NewReplacer("$GOOS", goos, "$EXT", ext, "$PWD", pwd)
	for src, dest := range copies {
		src = r.Replace(src)
		dest = r.Replace(dest)
		log.Printf("Copying %s to %s", src, dest)

		destDir := filepath.Dir(dest)
		os.MkdirAll(destDir, 0755)

		err := sh.Copy(dest, src)
		if err != nil {
			return err
		}
	}

	return SetBinExecutable()
}

// Run `chmod +x -R bin`.
func SetBinExecutable() error {
	err := chmodRecursive("bin", 0755)
	return errors.Wrap(err, "could not set +x on the test bin")
}

func chmodRecursive(name string, mode os.FileMode) error {
	return filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		log.Println("chmod +x ", path)
		return os.Chmod(path, mode)
	})
}

// Run integration tests (slow).
func TestIntegration() {
	mg.Deps(tests.EnsureTestCluster)

	var run string
	runTest := os.Getenv("PORTER_RUN_TEST")
	if runTest != "" {
		run = "-run=" + runTest
	}
	os.Setenv("GO111MODULE", "on")
	must.RunV("go", "build", "-o", "bin/testplugin", "./cmd/testplugin")
	must.Command("go", "test", "-timeout=30m", run, "-tags=integration", "./...").CollapseArgs().RunV()
}

func Install() {
	porterHome := getPorterHome()
	fmt.Println("installing Porter from bin to", porterHome)

	// Copy porter binaries
	mgx.Must(os.MkdirAll(porterHome, 0750))
	mgx.Must(shx.Copy(filepath.Join("bin", "porter"+xplat.FileExt()), porterHome))
	mgx.Must(shx.Copy(filepath.Join("bin", "runtimes"), porterHome, shx.CopyRecursive))

	// Copy mixin binaries
	mixinsDir := filepath.Join("bin", "mixins")
	mixinsDirItems, err := ioutil.ReadDir(mixinsDir)
	mgx.Must(errors.Wrap(err, "could not list mixins in bin"))
	for _, fi := range mixinsDirItems {
		if !fi.IsDir() {
			continue
		}

		mixin := fi.Name()
		srcDir := filepath.Join(mixinsDir, mixin)
		destDir := filepath.Join(porterHome, "mixins", mixin)
		mgx.Must(os.MkdirAll(destDir, 0750))

		// Copy the mixin client binary
		mgx.Must(shx.Copy(filepath.Join(srcDir, mixin+xplat.FileExt()), destDir))

		// Copy the mixin runtimes
		mgx.Must(shx.Copy(filepath.Join(srcDir, "runtimes"), destDir, shx.CopyRecursive))
	}
}

func getPorterHome() string {
	porterHome := os.Getenv("PORTER_HOME")
	if porterHome == "" {
		home, err := os.UserHomeDir()
		mgx.Must(errors.Wrap(err, "could not determine home directory"))
		porterHome = filepath.Join(home, ".porter")
	}
	return porterHome
}
