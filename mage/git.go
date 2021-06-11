package mage

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

// LoadMetadatda populates the status of the current working copy: current version, tag and permalink
func LoadMetadatda() GitMetadata {
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

// Get the name of the current branch, or the branch that contains the current tag
func getBranchName() string {
	// pull request
	if branch, ok := os.LookupEnv("SYSTEM_PULLREQUEST_SOURCEBRANCH"); ok {
		return branch
	}

	// branch build
	if branch, ok := os.LookupEnv("BUILD_SOURCEBRANCHNAME"); ok {
		return branch
	}

	// tag build
	// Use the first branch that contains the current commit
	matches, _ := must.OutputS("git", "for-each-ref", "--contains", "HEAD", "--format=%(refname:short)")
	firstMatchingBranch := strings.Split(matches, "\n")[0]
	// The matching branch may be a remote branch, just get its name
	return strings.Replace(firstMatchingBranch, "origin/", "", 1)
}

func getPermalink() (string, bool) {
	// Use latest for tagged commits
	taggedRelease := false
	permalinkSuffix := "canary"
	err := shx.RunS("git", "describe", "--tags", "--match=v*", "--exact")
	if err == nil {
		permalinkSuffix = "latest"
		taggedRelease = true
	}

	// Get the current branch name, or the name of the branch we tagged from
	branch := getBranchName()

	// Build a permalink such as "canary", "latest", "v1-latest", etc
	switch branch {
	case "main":
		return permalinkSuffix, taggedRelease
	default:
		return fmt.Sprintf("%s-%s", strings.TrimPrefix(branch, "release/"), permalinkSuffix), taggedRelease
	}
}
