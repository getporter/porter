package releases

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/mage/tools"
	"get.porter.sh/porter/pkg"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
	"github.com/pkg/errors"
)

var must = shx.CommandBuilder{StopOnError: true}

const (
	packagesRepo      = "bin/mixins/.packages"
	ReleaseRepository = "PORTER_RELEASE_REPOSITORY"
	PackagesRemote    = "PORTER_PACKAGES_REMOTE"
)

// Prepares bin directory for publishing a package
func preparePackageForPublish(pkgType string, name string) {
	info := LoadMetadata()

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
	mgx.Must(ioutil.WriteFile(askpass, []byte(contents), pkg.FileModeExecutable))

	pwd, _ := os.Getwd()
	script := filepath.Join(pwd, askpass)

	must.Command("git", "config", "core.askPass", script).In(dir).RunV()
}

func publishPackage(pkgType string, name string) {
	mg.Deps(tools.EnsureGitHubClient, ConfigureGitBot)

	info := LoadMetadata()

	repo := os.Getenv(ReleaseRepository)
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

	// Create or update GitHub release for the permalink (canary/latest) with the version's binaries
	if info.ShouldPublishPermalink() {
		// Move the permalink tag. The existing release automatically points to the tag.
		must.RunV("git", "tag", info.Permalink, info.Version+"^{}", "-f")
		must.RunV("git", "push", "-f", remote, info.Permalink)

		AddFilesToRelease(repo, info.Permalink, versionDir)
	} else {
		fmt.Println("Skipping publish package for permalink", info.Permalink)
	}

	// Create GitHub release for the exact version (v1.2.3) and attach assets
	if info.IsTaggedRelease {
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
	info := LoadMetadata()

	if !(info.Permalink == "canary" || info.IsTaggedRelease) {
		fmt.Println("Skipping publish package feed for permalink", info.Permalink)
		return
	}

	// Clone the packages repository
	if _, err := os.Stat(packagesRepo); !os.IsNotExist(err) {
		os.RemoveAll(packagesRepo)
	}
	remote := os.Getenv(PackagesRemote)
	if remote == "" {
		remote = fmt.Sprintf("https://github.com/getporter/packages.git")
	}
	must.RunV("git", "clone", "--depth=1", remote, packagesRepo)
	configureGitBotIn(packagesRepo)

	generatePackageFeed(pkgType)

	must.Command("git", "-c", "user.name='Porter Bot'", "-c", "user.email=bot@porter.sh", "commit", "--signoff", "-am", fmt.Sprintf("Add %s@%s to %s feed", name, info.Version, pkgType)).
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

	// Try to use a local copy of porter first, otherwise use the
	// one installed in GOAPTH/bin
	porterPath := "bin/porter"
	if _, err := os.Stat(porterPath); err != nil {
		porterPath = "porter"
		tools.EnsurePorter()
	}
	must.RunV(porterPath, "mixins", "feed", "generate", "-d", filepath.Join("bin", pkgDir), "-f", feedFile, "-t", "build/atom-template.xml")
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
func AddFilesToRelease(repo string, tag string, dir string) {
	files := listFiles(dir)

	checksumFiles := make([]string, len(files))
	for i, file := range files {
		checksumFile, added := AddChecksumExt(file)
		if !added {
			checksumFiles[i] = file
			continue
		}

		err := createChecksumFile(file, checksumFile)
		if err != nil {
			mgx.Must(errors.Wrapf(err, "failed to generate checksum file for asset %s", file))
		}
		checksumFiles[i] = checksumFile

	}

	files = append(files, checksumFiles...)

	// Mark canary releases as a draft
	draft := ""
	if strings.HasPrefix(tag, "canary") {
		draft = "-p"
	}

	if releaseExists(repo, tag) {
		must.Command("gh", "release", "upload", "--clobber", "-R", repo, tag).
			Args(files...).RunV()
	} else {
		must.Command("gh", "release", "create", "-R", repo, "-t", tag, "--notes=", draft, tag).
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

func AddChecksumExt(path string) (string, bool) {
	if filepath.Ext(path) == ".sha256sum" {
		return path, false
	}

	return path + ".sha256sum", true
}

func GenerateChecksum(data io.Reader, path string) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, data); err != nil {
		return "", err
	}
	sum := hash.Sum(nil)

	return AppendDataPath(sum, path), nil
}

func AppendDataPath(data []byte, path string) string {
	// write the checksum and file name to the checksum file so it can be
	// verified by tools like `shasum`
	return hex.EncodeToString(data) + "  " + filepath.Base(path)
}

func createChecksumFile(contentPath string, checksumFile string) error {
	data, err := os.Open(contentPath)
	if err != nil {
		return err
	}
	defer data.Close()

	sum, err := GenerateChecksum(data, contentPath)
	if err != nil {
		return err
	}

	f, err := os.Create(checksumFile)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(sum); err != nil {
		return err
	}

	return nil
}
