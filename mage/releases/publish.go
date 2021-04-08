package releases

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/pkg"
	"github.com/carolynvs/magex/pkg/archive"
	"github.com/carolynvs/magex/pkg/downloads"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
	"github.com/pkg/errors"
)

var must = shx.CommandBuilder{StopOnError: true}

const (
	packagesRepo = "bin/mixins/.packages"
)

// Prepares bin directory for publishing
func PrepareMixinForPublish(mixin string, version string, permalink string) {
	// Prepare the bin directory for generating a mixin feed
	// We want the bin to contain either a version directory (v1.2.3) or a canary directory.
	// We do not want a latest directory, latest entries are calculated using the most recent
	// timestamp in the atom.xml, not from an explicit entry.
	if permalink == "latest" {
		return
	}

	binDir := filepath.Join("bin/mixins/", mixin)
	// Temp hack until we have mixin.mk totally moved into mage
	if mixin == "porter" {
		binDir = "bin"
	}
	versionDir := filepath.Join(binDir, version)
	permalinkDir := filepath.Join(binDir, permalink)

	mgx.Must(os.RemoveAll(permalinkDir))
	log.Printf("mv %s %s\n", versionDir, permalinkDir)
	mgx.Must(shx.Copy(versionDir, permalinkDir, shx.CopyRecursive))
}

// Use GITHUB_TOKEN to log the porter bot into git
func ConfigureGitBot() {
	configureGitBotIn(".")
}

func configureGitBotIn(dir string) {
	pwd, _ := os.Getwd()
	script := filepath.Join(pwd, "build/git_askpass.sh")

	must.Command("git", "config", "user.name", "Porter Bot").In(dir).RunV()
	must.Command("git", "config", "user.email", "bot@porter.sh").In(dir).RunV()
	must.Command("git", "config", "core.askPass", script).In(dir).RunV()
}

// Publish a mixin's binaries.
func PublishMixin(mixin string, version string, permalink string) {
	mg.Deps(EnsureGitHubClient, ConfigureGitBot)

	repo := os.Getenv("PORTER_RELEASE_REPOSITORY")
	if repo == "" {
		repo = fmt.Sprintf("github.com/getporter/%s-mixin", mixin)
	}
	remote := fmt.Sprintf("https://%s.git", repo)
	versionDir := filepath.Join("bin/mixins/", mixin, version)

	// Move the permalink tag. The existing release automatically points to the tag.
	must.RunV("git", "tag", permalink, version+"^{}", "-f")
	must.RunV("git", "push", "-f", remote, permalink)

	// Create or update GitHub release for the permalink (canary/latest) with the version's binaries
	AddFilesToRelease(repo, permalink, versionDir)

	if permalink == "latest" {
		// Create GitHub release for the exact version (v1.2.3) and attach assets
		AddFilesToRelease(repo, version, versionDir)
	}
}

// Generate an updated mixin feed and publishes it.
func PublishMixinFeed(mixin string, version string) {
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

	GenerateMixinFeed()

	must.Command("git", "commit", "--signoff", "--author='Porter Bot<bot@porter.sh>'", "-am", fmt.Sprintf("Add %s@%s to mixin feed", mixin, version)).
		In(packagesRepo).RunV()
	must.Command("git", "push").In(packagesRepo).RunV()
}

// Generate a mixin feed from any mixin versions in bin/mixins.
func GenerateMixinFeed() {
	feedFile := filepath.Join(packagesRepo, "mixins/atom.xml")
	must.RunV("bin/porter", "mixins", "feed", "generate", "-d", "bin/mixins", "-f", feedFile, "-t", "build/atom-template.xml")
}

// Install the gh CLI
func EnsureGitHubClient() {
	if ok, _ := pkg.IsCommandAvailable("gh", ""); ok {
		return
	}

	// gh cli unfortunately uses a different archive schema depending on the OS
	target := "gh_{{.VERSION}}_{{.GOOS}}_{{.GOARCH}}/bin/gh{{.EXT}}"
	if runtime.GOOS == "windows" {
		target = "bin/gh.exe"
	}

	opts := archive.DownloadArchiveOptions{
		DownloadOptions: downloads.DownloadOptions{
			UrlTemplate: "https://github.com/cli/cli/releases/download/v{{.VERSION}}/gh_{{.VERSION}}_{{.GOOS}}_{{.GOARCH}}{{.EXT}}",
			Name:        "gh",
			Version:     "1.8.1",
			OsReplacement: map[string]string{
				"darwin": "macOS",
			},
		},
		ArchiveExtensions: map[string]string{
			"linux":   ".tar.gz",
			"darwin":  ".tar.gz",
			"windows": ".zip",
		},
		TargetFileTemplate: target,
	}

	err := archive.DownloadToGopathBin(opts)
	mgx.Must(err)
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
