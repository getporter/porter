package releases

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"get.porter.sh/porter/mage"
	"get.porter.sh/porter/mage/tools"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
	"github.com/pkg/errors"
)

var must = shx.CommandBuilder{StopOnError: true}

const (
	packagesRepo = "bin/mixins/.packages"
)

// Prepares bin directory for publishing a package
func preparePackageForPublish(pkgType string, name string) {
	info := mage.LoadMetadatda()

	// Prepare the bin directory for generating a package feed
	// We want the bin to contain either a version directory (v1.2.3) or a canary directory.
	// We do not want a latest directory, latest entries are calculated using the most recent
	// timestamp in the atom.xml, not from an explicit entry.
	if info.IsTaggedRelease {
		return
	}

	binDir := filepath.Join("bin", pkgType+"s", name)
	// Temp hack until we have mixin.mk totally moved into mage
	if name == "porter" {
		binDir = "bin"
	}
	versionDir := filepath.Join(binDir, info.Version)
	permalinkDir := filepath.Join(binDir, info.Permalink)

	mgx.Must(os.RemoveAll(permalinkDir))
	log.Printf("mv %s %s\n", versionDir, permalinkDir)
	mgx.Must(shx.Copy(versionDir, permalinkDir, shx.CopyRecursive))
}

// Prepares bin directory for publishing a mixin
func PrepareMixinForPublish(mixin string) {
	preparePackageForPublish("mixin", mixin)
}

// Prepares bin directory for publishing a plugin
func PreparePluginForPublish(plugin string) {
	preparePackageForPublish("plugin", plugin)
}

// Use GITHUB_TOKEN to log the porter bot into git
func ConfigureGitBot() {
	configureGitBotIn(".")
}

func configureGitBotIn(dir string) {
	askpass := "build/git_askpass.sh"
	contents := `#!/bin/sh
exec echo "$GITHUB_TOKEN"
`
	mgx.Must(ioutil.WriteFile(askpass, []byte(contents), 0755))

	pwd, _ := os.Getwd()
	script := filepath.Join(pwd, askpass)

	must.Command("git", "config", "user.name", "Porter Bot").In(dir).RunV()
	must.Command("git", "config", "user.email", "bot@porter.sh").In(dir).RunV()
	must.Command("git", "config", "core.askPass", script).In(dir).RunV()
}

func publishPackage(pkgType string, name string) {
	mg.Deps(tools.EnsureGitHubClient, ConfigureGitBot)

	info := mage.LoadMetadatda()

	repo := os.Getenv("PORTER_RELEASE_REPOSITORY")
	if repo == "" {
		switch pkgType {
		case "mixin":
			repo = fmt.Sprintf("github.com/getporter/%s-mixin", name)
		case "plugin":
			repo = fmt.Sprintf("github.com/getporter/%s-plugins", name)
		default:
			mgx.Must(errors.Errorf("invalid package type %q", pkgType))
		}
	}
	remote := fmt.Sprintf("https://%s.git", repo)
	versionDir := filepath.Join("bin", pkgType+"s", name, info.Version)

	// Move the permalink tag. The existing release automatically points to the tag.
	must.RunV("git", "tag", info.Permalink, info.Version+"^{}", "-f")
	must.RunV("git", "push", "-f", remote, info.Permalink)

	// Create or update GitHub release for the permalink (canary/latest) with the version's binaries
	AddFilesToRelease(repo, info.Permalink, versionDir)

	if info.IsTaggedRelease {
		// Create GitHub release for the exact version (v1.2.3) and attach assets
		AddFilesToRelease(repo, info.Version, versionDir)
	}
}

// Publish a mixin's binaries.
func PublishMixin(mixin string) {
	publishPackage("mixin", mixin)
}

// Publish a plugin's binaries.
func PublishPlugin(plugin string) {
	publishPackage("plugin", plugin)

}

func publishPackageFeed(pkgType string, name string) {
	info := mage.LoadMetadatda()

	// Clone the packages repository
	if _, err := os.Stat(packagesRepo); !os.IsNotExist(err) {
		os.RemoveAll(packagesRepo)
	}
	remote := os.Getenv("PORTER_PACKAGES_REMOTE")
	if remote == "" {
		remote = fmt.Sprintf("https://github.com/getporter/packages.git")
	}
	must.RunV("git", "clone", "--depth=1", remote, packagesRepo)
	configureGitBotIn(packagesRepo)

	generatePackageFeed(pkgType)

	must.Command("git", "commit", "--signoff", "--author='Porter Bot<bot@porter.sh>'", "-am", fmt.Sprintf("Add %s@%s to %s feed", name, info.Version, pkgType)).
		In(packagesRepo).RunV()
	must.Command("git", "push").In(packagesRepo).RunV()
}

// Generate an updated mixin feed and publishes it.
func PublishMixinFeed(mixin string) {
	publishPackageFeed("mixin", mixin)
}

// Generate an updated plugin feed and publishes it.
func PublishPluginFeed(plugin string) {
	publishPackageFeed("plugin", plugin)
}

func generatePackageFeed(pkgType string) {
	pkgDir := pkgType + "s"
	feedFile := filepath.Join(packagesRepo, pkgDir, "atom.xml")
	must.RunV("bin/porter", "mixins", "feed", "generate", "-d", filepath.Join("bin", pkgDir), "-f", feedFile, "-t", "build/atom-template.xml")
}

// Generate a mixin feed from any mixin versions in bin/mixins.
func GenerateMixinFeed() {
	generatePackageFeed("mixin")
}

// Generate a plugin feed from any plugin versions in bin/plugins.
func GeneratePluginFeed() {
	generatePackageFeed("plugin")
}

// AddFilesToRelease uploads the files in the specified directory to a GitHub release.
// If the release does not exist already, it will be created with empty release notes.
func AddFilesToRelease(repo string, version string, dir string) {
	files := listFiles(dir)

	// Mark canary releases as a draft
	draft := ""
	if version == "canary" {
		draft = "-p"
	}

	if releaseExists(repo, version) {
		must.Command("gh", "release", "upload", "--clobber", "-R", repo, version).
			Args(files...).RunV()
	} else {
		must.Command("gh", "release", "create", "-R", repo, "-t", version, "--notes=", draft, version).
			CollapseArgs().Args(files...).RunV()
	}
}

func releaseExists(repo string, version string) bool {
	return shx.RunE("gh", "release", "view", "-R", repo, version) == nil
}

func listFiles(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	mgx.Must(errors.Wrapf(err, "error listing files in %s", dir))

	names := make([]string, len(files))
	for i, fi := range files {
		names[i] = filepath.Join(dir, fi.Name())
	}

	return names
}
