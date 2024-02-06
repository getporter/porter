//go:build mage

// This is a magefile, and is a "makefile for go".
// See https://magefile.org/
package main

import (
	"bytes"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"get.porter.sh/magefiles/ci"
	"get.porter.sh/magefiles/git"
	"get.porter.sh/magefiles/releases"
	"get.porter.sh/magefiles/tools"
	"get.porter.sh/porter/mage/setup"
	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/tests/tester"
	mageci "github.com/carolynvs/magex/ci"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/carolynvs/magex/xplat"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"golang.org/x/sync/errgroup"

	// mage:import
	"get.porter.sh/magefiles/docker"

	// mage:import
	"get.porter.sh/magefiles/tests"

	// mage:import
	_ "get.porter.sh/porter/mage/docs"
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

func GenerateGRPCProtobufs() {
	mg.Deps(setup.EnsureBufBuild)
	must.Command("buf", "build").In("proto").RunV()
	must.Command("buf", "generate").In("proto").RunV()
}

// Builds all code artifacts in the repository
func Build() {
	mg.SerialDeps(InstallBuildTools, BuildPorter, BuildExecMixin, BuildAgent, DocsGen)
	mg.Deps(GetMixins)
}

func InstallBuildTools() {
	mg.Deps(setup.EnsureProtobufTools, setup.EnsureGRPCurl, setup.EnsureBufBuild)
}

// Build the porter client and runtime
func BuildPorter() {
	mg.Deps(Tidy, copySchema)

	mgx.Must(releases.BuildAll(PKG, "porter", "bin"))
}

// TODO: add support to decouple dir and command name to magefile repo
// TODO: add support for additional ldflags to magefile repo
// Build the porter client and runtime with gRPC server enabled
func XBuildPorterGRPCServer() {
	var g errgroup.Group
	supportedClientGOOS := []string{"linux", "darwin", "windows"}
	supportedClientGOARCH := []string{"amd64", "arm64"}
	srcCmd := "porter"
	srcPath := "./cmd/" + srcCmd
	outPath := "bin/dev"
	info := releases.LoadMetadata()
	ldflags := fmt.Sprintf("-w -X main.includeGRPCServer=true -X %s/pkg.Version=%s -X %s/pkg.Commit=%s", PKG, info.Version, PKG, info.Commit)
	os.MkdirAll(filepath.Dir(outPath), 0770)
	for _, goos := range supportedClientGOOS {
		goos := goos
		for _, goarch := range supportedClientGOARCH {
			goarch := goarch
			g.Go(func() error {
				cmdName := fmt.Sprintf("%s-api-server-%s-%s", srcCmd, goos, goarch)
				if goos == "windows" {
					cmdName = cmdName + ".exe"
				}
				out := filepath.Join(outPath, cmdName)
				return shx.Command("go", "build", "-ldflags", ldflags, "-o", out, srcPath).
					Env("CGO_ENABLED=0", "GO111MODULE=on", "GOOS="+goos, "GOARCH="+goarch).
					RunV()
			})
		}
	}
	mgx.Must(g.Wait())
}

func copySchema() {
	// Copy the porter manifest schema into our templates directory with the other schema
	// We can't use symbolic links because that doesn't work on windows
	mgx.Must(shx.Copy("pkg/schema/manifest.schema.json", "pkg/templates/templates/schema.json"))
	mgx.Must(shx.Copy("pkg/schema/manifest.v1.1.0.schema.json", "pkg/templates/templates/v1.1.0.schema.json"))
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

// Cross compile agent for multiple archs in linux
func XBuildAgent() {
	mg.Deps(BuildAgent)
	releases.XBuild(PKG, "agent", "bin", "linux", "arm64")
}

// Cross-compile porter and the exec mixin
func XBuildAll() {
	mg.Deps(XBuildPorter, XBuildMixins, XBuildAgent)
	mg.SerialDeps(XBuildPorterGRPCServer)
}

// Cross-compile porter
func XBuildPorter() {
	mg.Deps(copySchema)
	releases.XBuildAll(PKG, "porter", "bin")
	releases.PrepareMixinForPublish("porter")
}

// Cross-compile the exec mixin
func XBuildMixins() {
	releases.XBuildAll(PKG, "exec", "bin/mixins/exec")
	releases.PrepareMixinForPublish("exec")
}

// Generate cli documentation for the website
func DocsGen() {
	// Remove the generated cli directory so that it can detect deleted files
	os.RemoveAll("docs/content/cli")
	os.Mkdir("docs/content/cli", pkg.FileModeDirectory)

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
		{name: "helm3", feed: "https://mchorfa.github.io/porter-helm3/atom.xml", version: "v0.1.16"},
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

// Update golden test files (unit tests only) to match the new test outputs and re-run the unit tests
func UpdateTestfiles() {
	os.Setenv("PORTER_UPDATE_TEST_FILES", "true")
	defer os.Unsetenv("PORTER_UPDATE_TEST_FILES")

	// Run tests and update any golden files
	TestUnit()

	// Re-run the tests with the golden files locked in to make sure everything passes now
	os.Unsetenv("PORTER_UPDATE_TEST_FILES")
	TestUnit()
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

	TestInitWarnings()
}

// Run smoke tests to quickly check if Porter is broken
func TestSmoke() error {
	mg.Deps(copySchema, TryRegisterLocalHostAlias, docker.RestartDockerRegistry, BuildTestMixin)

	// Only do verbose output of tests when called with `mage -v TestSmoke`
	v := ""
	if mg.Verbose() {
		v = "-v"
	}

	// Adding -count to prevent go from caching the test results.
	return shx.Command("go", "test", "-count=1", "-timeout=20m", "-tags", "smoke", v, "./tests/smoke/...").CollapseArgs().RunV()
}

// Run grpc service tests
func TestGRPCService() {
	var run string
	runTest := os.Getenv("PORTER_RUN_TEST")
	if runTest != "" {
		run = "-run=" + runTest
	}

	verbose := ""
	if mg.Verbose() {
		verbose = "-v"
	}
	must.Command("go", "test", verbose, "-timeout=5m", run, "./tests/grpc/...").CollapseArgs().RunV()
}

func getRegistry() string {
	registry := os.Getenv("PORTER_REGISTRY")
	if registry == "" {
		registry = "localhost:5000"
	}
	return registry
}

func BuildImages() {
	info := releases.LoadMetadata()
	registry := getRegistry()

	buildImages(registry, info)
}

func buildGRPCProtocImage() {
	var g errgroup.Group

	enableBuildKit := "DOCKER_BUILDKIT=1"
	g.Go(func() error {
		img := "protoc:local"
		err := shx.Command("docker", "build", "-t", img, "-f", "build/protoc.Dockerfile", ".").
			Env(enableBuildKit).RunV()
		if err != nil {
			return err
		}
		return nil
	})
	mgx.Must(g.Wait())
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

func PublishServerMultiArchImages() {
	registry := getRegistry()
	info := releases.LoadMetadata()
	buildAndPushServerMultiArch(registry, info)
}

func buildAndPushServerMultiArch(registry string, info releases.GitMetadata) {
	img := fmt.Sprintf("%s/server:%s", registry, info.Version)
	must.RunV("docker", "buildx", "create", "--use")
	must.RunV("docker", "buildx", "bake", "-f", "docker-bake.json", "--push", "--set", "server.tags="+img, "server")
}

// Build a local image for the server based off of local architecture
func BuildLocalServerImage() {
	registry := getRegistry()
	info := releases.LoadMetadata()
	goarch := runtime.GOARCH
	buildServerImage(registry, info, goarch)
}

// Builds an image for the server based off of the goarch
func buildServerImage(registry string, info releases.GitMetadata, goarch string) {
	var platform string
	switch goarch {
	case "arm64":
		platform = "linux/arm64"
	case "amd64":
		platform = "linux/amd64"
	default:
		platform = "linux/amd64"
	}
	img := fmt.Sprintf("%s/server:%s", registry, info.Version)
	must.RunV("docker", "build", "-f", "build/images/server/Dockerfile", "-t", img, "--platform="+platform, ".")
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
	// Rewrites the version number in the script uploaded to the github release
	// If it's a tagged version, we reference that in the script
	// Otherwise reference the name of the build, e.g. "canary"
	scriptVersion := info.Version
	if !info.IsTaggedRelease {
		scriptVersion = info.Permalink
	}
	must.Command("./scripts/prep-install-scripts.sh").Env("VERSION=" + scriptVersion).RunV()

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
		must.RunV("git", "tag", "-f", info.Permalink, info.Version)
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
	if err != nil {
		return fmt.Errorf("could not set +x on the test bin: %w", err)
	}

	return nil
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
	mg.Deps(tests.EnsureTestCluster, copySchema, TryRegisterLocalHostAlias, BuildTestMixin, BuildTestPlugin)

	var run string
	runTest := os.Getenv("PORTER_RUN_TEST")
	if runTest != "" {
		run = "-run=" + runTest
	}

	verbose := ""
	if mg.Verbose() {
		verbose = "-v"
	}

	var path string
	filename := os.Getenv("PORTER_INTEG_FILE")
	if filename == "" {
		path = "./..."
	} else {
		path = "./tests/integration/" + filename
	}

	must.Command("go", "test", verbose, "-timeout=30m", run, "-tags=integration", path).CollapseArgs().RunV()
}

func TestInitWarnings() {
	// This is hard to test in a normal unit test because we need to build porter with custom build tags,
	// so I'm testing it in the magefile directly in a way that doesn't leave around an unsafe Porter binary.

	// Verify that running Porter with traceSensitiveAttributes set that a warning is printed
	fmt.Println("Validating traceSensitiveAttributes warning")
	output := &bytes.Buffer{}
	must.Command("go", "run", "-tags=traceSensitiveAttributes", "./cmd/porter", "schema").
		Stderr(output).Stdout(output).Exec()
	if !strings.Contains(output.String(), "WARNING! This is a custom developer build of Porter with the traceSensitiveAttributes build flag set") {
		fmt.Printf("Got output: %s\n", output.String())
		panic("Expected a build of Porter with traceSensitiveAttributes build tag set to warn at startup but it didn't")
	}
}

// TryRegisterLocalHostAlias edits /etc/hosts to use porter-test-registry hostname alias
// This is not safe to call more than once and is intended for use on the CI server only
func TryRegisterLocalHostAlias() {
	if _, isCI := mageci.DetectBuildProvider(); !isCI {
		return
	}

	err := shx.RunV("sudo", "bash", "-c", "echo 127.0.0.1 porter-test-registry >> /etc/hosts")
	if err != nil {
		fmt.Println("skipping registering the porter-test-registry hostname alias: could not write to /etc/hosts")
		return
	}

	fmt.Println("Added host alias porter-test-registry to /etc/hosts")
	os.Setenv(tester.TestRegistryAlias, "porter-test-registry")
}

func BuildTestPlugin() {
	must.RunV("go", "build", "-o", "bin/testplugin", "./cmd/testplugin")
}

func BuildTestMixin() {
	os.MkdirAll("bin/mixins/testmixin", 0770)
	must.RunV("go", "build", "-o", "bin/mixins/testmixin/testmixin"+xplat.FileExt(), "./cmd/testmixin")
}

// Copy the locally built porter and exec binaries to PORTER_HOME
func Install() {
	porterHome := getPorterHome()
	fmt.Println("installing Porter from bin to", porterHome)

	// Copy porter binaries
	mgx.Must(os.MkdirAll(porterHome, pkg.FileModeDirectory))

	// HACK: Works around a codesigning problem on Apple Silicon where overwriting a binary that has already been executed doesn't cause the corresponding codesign entry in the OS cache to update
	// Mac then prevents the updated binary from running because the signature doesn't match
	// Removing the file first clears the cache so that we don't run into "zsh: killed porter..."
	// See https://stackoverflow.com/questions/67378106/mac-m1-cping-binary-over-another-results-in-crash
	// See https://openradar.appspot.com/FB8914231
	mgx.Must(os.Remove(filepath.Join(porterHome, "porter"+xplat.FileExt())))
	mgx.Must(os.RemoveAll(filepath.Join(porterHome, "runtimes")))

	// Okay now it's safe to copy these files over
	mgx.Must(shx.Copy(filepath.Join("bin", "porter"+xplat.FileExt()), porterHome))
	mgx.Must(shx.Copy(filepath.Join("bin", "runtimes"), porterHome, shx.CopyRecursive))

	// Copy mixin binaries
	mixinsDir := filepath.Join("bin", "mixins")
	mixinsDirItems, err := os.ReadDir(mixinsDir)
	if err != nil {
		mgx.Must(fmt.Errorf("could not list mixins in bin: %w", err))
	}

	for _, fi := range mixinsDirItems {
		// do not install the test mixins
		if fi.Name() == "testmixin" {
			continue
		}

		if !fi.IsDir() {
			continue
		}

		mixin := fi.Name()
		srcDir := filepath.Join(mixinsDir, mixin)
		destDir := filepath.Join(porterHome, "mixins", mixin)
		mgx.Must(os.MkdirAll(destDir, pkg.FileModeDirectory))

		// HACK: Works around a codesigning problem on Apple Silicon where overwriting a binary that has already been executed doesn't cause the corresponding codesign entry in the OS cache to update
		// Mac then prevents the updated binary from running because the signature doesn't match
		// Removing the file first clears the cache so that we don't run into "zsh: killed MIXIN..."
		// See https://stackoverflow.com/questions/67378106/mac-m1-cping-binary-over-another-results-in-crash
		// See https://openradar.appspot.com/FB8914231
		mgx.Must(os.Remove(filepath.Join(destDir, mixin+xplat.FileExt())))
		mgx.Must(os.RemoveAll(filepath.Join(destDir, "runtimes")))

		// Copy the mixin client binary
		mgx.Must(shx.Copy(filepath.Join(srcDir, mixin+xplat.FileExt()), destDir))

		// Copy the mixin runtimes
		mgx.Must(shx.Copy(filepath.Join(srcDir, "runtimes"), destDir, shx.CopyRecursive))
	}
}

// Run Go Vet on the project
func Vet() {
	must.RunV("go", "vet", "./...")
}

// Run staticcheck on the project
func Lint() {
	mg.Deps(tools.EnsureStaticCheck)
	must.RunV("staticcheck", "./...")
}

func getPorterHome() string {
	porterHome := os.Getenv("PORTER_HOME")
	if porterHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			mgx.Must(fmt.Errorf("could not determine home directory: %w", err))
		}

		porterHome = filepath.Join(home, ".porter")
	}
	return porterHome
}

// SetupDCO configures your git repository to automatically sign your commits
// to comply with our DCO
func SetupDCO() error {
	return git.SetupDCO()
}
