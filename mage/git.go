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
}

// LoadMetadatda populates the status of the current working copy: current version, tag and permalink
func LoadMetadatda() GitMetadata {
	loadMetadata.Do(func() {
		gitMetadata = GitMetadata{}

		// Get a description of the commit, e.g. v0.30.1 (latest) or v0.30.1-32-gfe72ff73 (canary)
		version, _ := must.OutputS("git", "describe", "--tags", "--match=v*")
		if version == "" {
			// repo without any tags in it
			gitMetadata.Version = "v0.0.0"
		} else {
			gitMetadata.Version = version
		}

		// Get the hash for the current commit
		gitMetadata.Commit, _ = must.OutputS("git", "rev-parse", "--short", "HEAD")

		// Use latest for tagged commits
		permalinkSuffix := "canary"
		err := shx.RunS("git", "describe", "--tags", "--match=v*", "--exact")
		if err == nil {
			permalinkSuffix = "latest"
		}

		// Get the current branch name, or the name of the branch we tagged from
		branch, ok := os.LookupEnv("BUILD_SOURCEBRANCHNAME") // Azure pipelines doesn't build from the branch (they do a detached head for branch builds) but we can get it from the environment
		if !ok {
			// Use the first branch that contains the current commit
			matches, _ := must.OutputS("git", "for-each-ref", "--contains", "HEAD", "--format=%(refname:short)")
			branch = strings.Replace(strings.Split(matches, "\n")[0], "origin/", "", 1)
		}

		// Build a permalink such as "canary", "latest", "v1-latest", etc
		switch branch {
		case "", "main":
			gitMetadata.Permalink = permalinkSuffix
		default:
			gitMetadata.Permalink = fmt.Sprintf("%s-%s", strings.TrimPrefix(branch, "release/"), permalinkSuffix)
		}
	})

	log.Println("Permalink:", gitMetadata.Permalink)
	log.Println("Version:", gitMetadata.Version)
	log.Println("Commit:", gitMetadata.Commit)

	if githubEnv, ok := os.LookupEnv("GITHUB_ENV"); ok {
		err := ioutil.WriteFile(githubEnv, []byte("PERMALINK="+gitMetadata.Permalink), 0644)
		mgx.Must(errors.Wrapf(err, "couldn't persist PERMALINK to a GitHub Actions environment variable"))
	}

	return gitMetadata
}
