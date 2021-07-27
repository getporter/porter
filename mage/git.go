package mage

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/pkg/errors"
)

var gitMetadata GitMetadata
var loadMetadata sync.Once
var must = shx.CommandBuilder{StopOnError: true}

type GitMetadata struct {
	// Permalink is the version alias, e.g. latest, or canary
	Permalink string

	// Version is the tag or tag+commit hash
	Version string

	// Commit is the hash of the current commit
	Commit string

	// IsTaggedRelease indicates if the build is for a versioned tag
	IsTaggedRelease bool
}

func (m GitMetadata) ShouldPublishPermalink() bool {
	// For now don't publish canary-v1 or latest-v1 to keep things simpler
	return m.Permalink == "canary" || m.Permalink == "latest"
}

// LoadMetadata populates the status of the current working copy: current version, tag and permalink
func LoadMetadata() GitMetadata {
	loadMetadata.Do(func() {
		gitMetadata = GitMetadata{
			Version: getVersion(),
			Commit:  getCommit(),
		}

		gitMetadata.Permalink, gitMetadata.IsTaggedRelease = getPermalink()

		log.Println("Tagged Release:", gitMetadata.IsTaggedRelease)
		log.Println("Permalink:", gitMetadata.Permalink)
		log.Println("Version:", gitMetadata.Version)
		log.Println("Commit:", gitMetadata.Commit)
	})

	// Save github action environment variables
	if githubEnv, ok := os.LookupEnv("GITHUB_ENV"); ok {
		err := ioutil.WriteFile(githubEnv, []byte("PERMALINK="+gitMetadata.Permalink), 0644)
		mgx.Must(errors.Wrapf(err, "couldn't persist PERMALINK to a GitHub Actions environment variable"))
	}

	return gitMetadata
}

// Get the hash of the current commit
func getCommit() string {
	commit, _ := must.OutputS("git", "rev-parse", "--short", "HEAD")
	return commit
}

// Get a description of the commit, e.g. v0.30.1 (latest) or v0.30.1-32-gfe72ff73 (canary)
func getVersion() string {
	version, _ := must.OutputS("git", "describe", "--tags", "--match=v*")
	if version != "" {
		return version
	}

	// repo without any tags in it
	return "v0.0.0"
}

// Return either "main", "v*", or "dev" for all other branches.
func getBranchName() string {
	gitOutput, _ := must.OutputS("git", "for-each-ref", "--contains", "HEAD", "--format=%(refname)")
	refs := strings.Split(gitOutput, "\n")

	return pickBranchName(refs)
}

// Return either "main", "v*", or "dev" for all other branches.
func pickBranchName(refs []string) string {
	var branch string

	if b, ok := os.LookupEnv("SYSTEM_PULLREQUEST_SOURCEBRANCH"); ok {
		// pull request
		branch = b
	} else if b, ok := os.LookupEnv("BUILD_SOURCEBRANCH"); ok && !strings.HasPrefix(b, "refs/tags/") {
		// branch build
		// BUILD_SOURCEBRANCHNAME has the short name, e.g. main. BUILD_SOURCEBRANCH has the full name, e.g. refs/heads/main
		// They are populated for both tags and branches
		branch = os.Getenv("BUILD_SOURCEBRANCHNAME")
	} else {
		// tag build
		// Detect if this was a tag on main or a release
		sort.Strings(refs) // put main ahead of release/v*
		for _, ref := range refs {
			// Ignore tags
			if strings.HasSuffix(ref, "refs/tags") {
				continue
			}

			// Only match main and release/v* branches
			if strings.HasSuffix(ref, "/main") || strings.Contains(ref, "/release/v") {
				branch = ref
				break
			}
		}
	}

	// Convert the ref name into a branch name, e.g. refs/heads/main -> main
	branch = strings.NewReplacer("refs/heads/", "", "refs/remotes/origin/", "").Replace(branch)

	// Only use the following branch names "main", "release/v*", and "dev" for everything else
	if branch != "main" && !strings.HasPrefix(branch, "release/v") {
		branch = "dev"
	}

	// Convert release/v1 -> v1
	branch = strings.ReplaceAll(branch, "release/", "")
	return branch
}

func getPermalink() (string, bool) {
	// Use latest for tagged commits
	taggedRelease := false
	permalinkPrefix := "canary"
	err := shx.RunS("git", "describe", "--tags", "--match=v*", "--exact")
	if err == nil {
		permalinkPrefix = "latest"
		taggedRelease = true
	}

	// Get the current branch name, or the name of the branch we tagged from
	branch := getBranchName()

	// Build a permalink such as "canary", "latest", "latest-v1", or "dev-canary"
	switch branch {
	case "main":
		return permalinkPrefix, taggedRelease
	default:
		return fmt.Sprintf("%s-%s", permalinkPrefix, strings.TrimPrefix(branch, "release/")), taggedRelease
	}
}
