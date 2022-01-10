package mixins

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/carolynvs/magex/mgx"

	"get.porter.sh/porter/mage/releases"
	"get.porter.sh/porter/mage/tools"
	"github.com/carolynvs/magex/shx"
	"github.com/carolynvs/magex/xplat"
	"github.com/magefile/mage/mg"
)

type Magefile struct {
	Pkg       string
	MixinName string
	BinDir    string
}

// Create a magefile helper for a mixin
func NewMagefile(pkg, mixinName, binDir string) Magefile {
	return Magefile{Pkg: pkg, MixinName: mixinName, BinDir: binDir}
}

var must = shx.CommandBuilder{StopOnError: true}

// Build the mixin
func (m Magefile) Build() {
	must.RunV("go", "mod", "tidy")
	releases.BuildAll(m.Pkg, m.MixinName, m.BinDir)
}

// Cross-compile the mixin before a release
func (m Magefile) XBuildAll() {
	releases.XBuildAll(m.Pkg, m.MixinName, m.BinDir)
}

// Run unit tests
func (m Magefile) TestUnit() {
	v := ""
	if mg.Verbose() {
		v = "-v"
	}
	must.Command("go", "test", v, "./pkg/...").CollapseArgs().RunV()
}

// Run all tests
func (m Magefile) Test() {
	m.TestUnit()

	// Check that we can call `mixin version`
	m.Build()
	must.RunV(filepath.Join(m.BinDir, m.MixinName+xplat.FileExt()), "version")
}

// Publish the mixin and its mixin feed
func (m Magefile) Publish() {
	mg.SerialDeps(m.PublishBinaries, m.PublishMixinFeed)
}

// Publish binaries to a github release
// Requires PORTER_RELEASE_REPOSITORY to be set to github.com/USERNAME/REPO
func (m Magefile) PublishBinaries() {
	releases.PrepareMixinForPublish(m.MixinName)
	releases.PublishMixin(m.MixinName)
}

// Publish a mixin feed
// Requires PORTER_PACKAGES_REMOTE to be set to git@github.com:USERNAME/REPO.git
func (m Magefile) PublishMixinFeed() {
	mg.Deps(tools.EnsurePorter)
	releases.PublishMixinFeed(m.MixinName)
}

// Test out publish locally, with your github forks
// Assumes that you forked and kept the repository name unchanged.
func (m Magefile) TestPublish(username string) {
	// Backup the git config file, publish will set the user to a bot
	mgx.Must(shx.Copy(".git/config", ".git/config.bak"))

	os.Setenv(releases.ReleaseRepository, fmt.Sprintf("github.com/%s/%s-mixin", username, m.MixinName))
	os.Setenv(releases.PackagesRemote, fmt.Sprintf("https://github.com/%s/packages.git", username))

	m.Publish()

	// Restore the original git config
	os.Remove(".git/config")
	os.Rename(".git/config.bak", ".git/config")
}

// Install the mixin
func (m Magefile) Install() {
	porterHome := os.Getenv("PORTER_HOME")
	if porterHome == "" {
		home, _ := os.UserHomeDir()
		porterHome = filepath.Join(home, ".porter")
	}
	if _, err := os.Stat(porterHome); err != nil {
		panic("Could not find a Porter installation. Make sure that Porter is installed and set PORTER_HOME if you are using a non-standard installation path")
	}
	fmt.Printf("Installing the %s mixin into %s\n", m.MixinName, porterHome)

	os.MkdirAll(filepath.Join(porterHome, "mixins", m.MixinName, "runtimes"), 0700)
	mgx.Must(shx.Copy(filepath.Join(m.BinDir, m.MixinName+xplat.FileExt()), filepath.Join(porterHome, "mixins", m.MixinName)))
	mgx.Must(shx.Copy(filepath.Join(m.BinDir, "runtimes", m.MixinName+"-runtime"+xplat.FileExt()), filepath.Join(porterHome, "mixins/runtimes")))
}

// Remove generated build files
func (m Magefile) Clean() {
	os.RemoveAll("bin")
}
