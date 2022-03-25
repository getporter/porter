//go:build mage
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

	"get.porter.sh/magefiles/ci"
	// mage:import
	"get.porter.sh/magefiles/tests"
	// mage:import
	_ "get.porter.sh/porter/mage/docs"

	"get.porter.sh/magefiles/docker"
	"get.porter.sh/magefiles/releases"
	"get.porter.sh/magefiles/tools"
	"get.porter.sh/porter/pkg"
	"github.com/carolynvs/magex/mgx"
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
	PKG       = "get.porter.sh/porter"
	GoVersion = ">=1.17"
)

var must = shx.CommandBuilder{StopOnError: true}

// Check if we have the right version of Go
func CheckGoVersion() {
	tools.EnforceGoVersion(GoVersion)
}

// Builds all code artifacts in the repository
func Build() {
	mg.SerialDeps(BuildPorter, DocsGen, BuildExecMixin, BuildAgent)
	mg.Deps(GetMixins)
}

// Build the porter client and runtime
func BuildPorter() {
	mg.Deps(Tidy, copySchema)

	mgx.Must(releases.BuildAll(PKG, "porter", "bin"))
}

func copySchema() {
	// Copy the porter manifest schema into our templates directory with the other schema
	// We can't use symbolic links because that doesn't work on windows
	mgx.Must(shx.Copy("pkg/schema/manifest.schema.json", "pkg/templates/templates/schema.json"))
}

func Tidy() error {
	return shx.Run("go", "mod", "tidy")
}

// Build the exec mixin client and runtime
func BuildExecMixin() {
	mgx.Must(releases.BuildAll(PKG, "exec", "bin/mixins/exec"))
}

// Build the porter agent
func BuildAgent() {
	// the agent is only used embedded in a docker container, so we only build for linux
	releases.XBuild(PKG, "agent", "bin", "linux", "amd64")
}

// Cross-compile porter and the exec mixin
func XBuildAll() {
	mg.Deps(XBuildPorter, XBuildMixins, BuildAgent)
}

// Cross-compile porter
func XBuildPorter() {
	mg.Deps(copySchema)
	releases.XBuildAll(PKG, "porter", "bin")
}

// Cross-compile the exec mixin
func XBuildMixins() {
	releases.XBuildAll(PKG, "exec", "bin/mixins/exec")
}

// Generate cli documentation for the website
func DocsGen() {
	must.RunV("go", "run", "--tags=docs", "./cmd/porter", "docs")
}

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
	releases.LoadMetadata()
}

// ConfigureAgent sets up an Azure DevOps agent with EnsureMage and ensures
// that GOPATH/bin is in PATH.
func ConfigureAgent() error {
	return ci.ConfigureAgent()
}

// Install mixins used by tests and example bundles, if not already installed
func GetMixins() error {
	defaultMixinVersion := os.Getenv("MIXIN_TAG")
	if defaultMixinVersion == "" {
		defaultMixinVersion = "canary"
	}

	mixins := []struct {
		name    string
		url     string
		feed    string
		version string
	}{
		{name: "docker"},
		{name: "docker-compose"},
		{name: "arm"},
		{name: "terraform"},
		{name: "kubernetes"},
		{name: "helm3", feed: "https://mchorfa.github.io/porter-helm3/atom.xml", version: "v0.1.14"},
	}
	var errG errgroup.Group
	for _, mixin := range mixins {
		mixin := mixin
		mixinDir := filepath.Join("bin/mixins/", mixin.name)
		if _, err := os.Stat(mixinDir); err == nil {
			log.Println("Mixin already installed into bin:", mixin.name)
			continue
		}

		errG.Go(func() error {
			log.Println("Installing mixin:", mixin.name)
			if mixin.version == "" {
				mixin.version = defaultMixinVersion
			}
			var source string
			if mixin.feed != "" {
				source = "--feed-url=" + mixin.feed
			} else {
				source = "--url=" + mixin.url
			}
			return porter("mixin", "install", mixin.name, "--version", mixin.version, source).Run()
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
	must.Command("go", "test", "./...").Env("PORTER_UPDATE_TEST_FILES=true").RunV()
	must.RunV("make", "test-unit")
}

// Run all tests known to human-kind
func Test() {
	mg.Deps(TestUnit, TestSmoke, TestIntegration)
}

// Run unit tests and verify integration tests compile
func TestUnit() {
	mg.Deps(copySchema)

	// Only do verbose output of tests when called with `mage -v TestSmoke`
	v := ""
	if mg.Verbose() {
		v = "-v"
	}

	must.Command("go", "test", v, "./...").CollapseArgs().RunV()

	// Verify integration tests compile since we don't run them automatically on pull requests
	must.Run("go", "test", "-run=non", "-tags=integration", "./...")
}

// Run smoke tests to quickly check if Porter is broken
func TestSmoke() error {
	mg.Deps(copySchema)

	mg.Deps(docker.RestartDockerRegistry)

	// Only do verbose output of tests when called with `mage -v TestSmoke`
	v := ""
	if mg.Verbose() {
		v = "-v"
	}

	// Adding -count to prevent go from caching the test results.
	return shx.Command("go", "test", "-count=1", "-tags", "smoke", v, "./tests/smoke/...").CollapseArgs().RunV()
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
	info := releases.LoadMetadata()
	registry := getRegistry()

	buildImages(registry, info)
	if getDualPublish() {
		buildImages("ghcr.io/getporter", info)
	}
}

func buildImages(registry string, info releases.GitMetadata) {
	var g errgroup.Group

	enableBuildKit := "DOCKER_BUILDKIT=1"
	g.Go(func() error {
		img := fmt.Sprintf("%s/porter:%s", registry, info.Version)
		err := shx.Command("docker", "build", "-t", img, "-f", "build/images/client/Dockerfile", ".").
			Env(enableBuildKit).RunV()
		if err != nil {
			return err
		}

		err = shx.Run("docker", "tag", img, fmt.Sprintf("%s/porter:%s", registry, info.Permalink))
		if err != nil {
			return err
		}

		// porter-agent does a FROM porter so they can't go in parallel
		img = fmt.Sprintf("%s/porter-agent:%s", registry, info.Version)
		err = shx.Command("docker", "build", "-t", img, "--build-arg", "PORTER_VERSION="+info.Version, "--build-arg", "REGISTRY="+registry, "-f", "build/images/agent/Dockerfile", ".").
			Env(enableBuildKit).RunV()
		if err != nil {
			return err
		}

		return shx.Run("docker", "tag", img, fmt.Sprintf("%s/porter-agent:%s", registry, info.Permalink))
	})

	g.Go(func() error {
		img := fmt.Sprintf("%s/workshop:%s", registry, info.Version)
		err := shx.Command("docker", "build", "-t", img, "-f", "build/images/workshop/Dockerfile", ".").
			Env(enableBuildKit).RunV()
		if err != nil {
			return err
		}

		return shx.Run("docker", "tag", img, fmt.Sprintf("%s/workshop:%s", registry, info.Permalink))
	})

	mgx.Must(g.Wait())
}

func PublishImages() {
	mg.Deps(BuildImages)

	info := releases.LoadMetadata()

	pushImagesTo(getRegistry(), info)
	if getDualPublish() {
		pushImagesTo("ghcr.io/getporter", info)
	}
}

// Builds the porter-agent image and publishes it to a local test cluster with the Porter Operator.
func LocalPorterAgentBuild() {
	// Publish to the local registry/cluster setup by the Porter Operator.
	os.Setenv("REGISTRY", "localhost:5000")
	// Force the image to be pushed to the registry even though it's a local dev build.
	os.Setenv("PORTER_FORCE_PUBLISH", "true")

	mg.SerialDeps(XBuildPorter, BuildAgent, PublishImages)
}

// Only push tagged versions, canary and latest
func pushImagesTo(registry string, info releases.GitMetadata) {
	if info.IsTaggedRelease {
		pushImages(registry, info.Version)
	}

	force, _ := strconv.ParseBool(os.Getenv("PORTER_FORCE_PUBLISH"))
	if info.ShouldPublishPermalink() || force {
		pushImages(registry, info.Permalink)
	} else {
		fmt.Println("Skipping image publish for permalink", info.Permalink)
	}
}

func pushImages(registry string, tag string) {
	pushImage(fmt.Sprintf("%s/porter:%s", registry, tag))
	pushImage(fmt.Sprintf("%s/porter-agent:%s", registry, tag))
	pushImage(fmt.Sprintf("%s/workshop:%s", registry, tag))
}

func pushImage(img string) {
	must.RunV("docker", "push", img)
}

// Publish the porter binaries and install scripts.
func PublishPorter() {
	mg.Deps(tools.EnsureGitHubClient, releases.ConfigureGitBot)

	info := releases.LoadMetadata()

	// Copy install scripts into version directory
	must.Command("./scripts/prep-install-scripts.sh").Env("VERSION=" + info.Version).RunV()

	porterVersionDir := filepath.Join("bin", info.Version)
	execVersionDir := filepath.Join("bin/mixins/exec", info.Version)
	var repo = os.Getenv("PORTER_RELEASE_REPOSITORY")
	if repo == "" {
		repo = "github.com/getporter/porter"
	}
	remote := fmt.Sprintf("https://%s.git", repo)

	// Create or update GitHub release for the permalink (canary/latest) with the version's assets (porter binaries, exec binaries and install scripts)
	if info.ShouldPublishPermalink() {
		// Move the permalink tag. The existing release automatically points to the tag.
		must.RunV("git", "tag", info.Permalink, info.Version+"^{}", "-f")
		must.RunV("git", "push", "-f", remote, info.Permalink)

		releases.AddFilesToRelease(repo, info.Permalink, porterVersionDir)
		releases.AddFilesToRelease(repo, info.Permalink, execVersionDir)
	} else {
		fmt.Println("Skipping publish binaries for permalink", info.Permalink)
	}

	if info.IsTaggedRelease {
		// Create GitHub release for the exact version (v1.2.3) and attach assets
		releases.AddFilesToRelease(repo, info.Version, porterVersionDir)
		releases.AddFilesToRelease(repo, info.Version, execVersionDir)
	} else {
		fmt.Println("Skipping publish binaries for not tagged release", info.Version)
	}
}

// Publish internal porter mixins, like exec.
func PublishMixins() {
	releases.PublishMixinFeed("exec")
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
		os.MkdirAll(destDir, pkg.FileModeDirectory)

		err := sh.Copy(dest, src)
		if err != nil {
			return err
		}
	}

	return SetBinExecutable()
}

// Run `chmod +x -R bin`.
func SetBinExecutable() error {
	err := chmodRecursive("bin", pkg.FileModeExecutable)
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
	mg.Deps(tests.EnsureTestCluster, copySchema)

	var run string
	runTest := os.Getenv("PORTER_RUN_TEST")
	if runTest != "" {
		run = "-run=" + runTest
	}

	verbose := ""
	if mg.Verbose() {
		verbose = "-v"
	}
	must.RunV("go", "build", "-o", "bin/testplugin", "./cmd/testplugin")
	must.Command("go", "test", verbose, "-timeout=30m", run, "-tags=integration", "./...").CollapseArgs().RunV()
}

// Copy the locally built porter and exec binaries to PORTER_HOME
func Install() {
	porterHome := getPorterHome()
	fmt.Println("installing Porter from bin to", porterHome)

	// Copy porter binaries
	mgx.Must(os.MkdirAll(porterHome, pkg.FileModeDirectory))
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
		mgx.Must(os.MkdirAll(destDir, pkg.FileModeDirectory))

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
