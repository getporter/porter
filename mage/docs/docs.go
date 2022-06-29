package docs

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/magefiles/docker"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
)

var must = shx.CommandBuilder{StopOnError: true}

const (
	LocalOperatorRepositoryEnv = "PORTER_OPERATOR_REPOSITORY"
	PreviewContainer           = "porter-docs"

	// DefaultOperatorSourceDir is the directory where the Porter Operator docs
	// are cloned when LocalOperatorRepositoryEnv was not specified.
	DefaultOperatorSourceDir = "docs/sources/operator"
)

// Build the website in preparation for deploying to the production website on Netlify.
// Uses symlinks so it won't work on Windows. Don't run locally, use DocsPreview instead.
func Docs() {
	// Remove the preview container because otherwise it holds a file open and we can't delete the volume mount created at docs/content/operator
	mg.SerialDeps(removePreviewContainer, linkOperatorDocs)

	cmd := must.Command("hugo", "--source", "docs/")
	baseURL := os.Getenv("BASEURL")
	if baseURL != "" {
		cmd.Args("-b", baseURL)
	}
	cmd.RunV()
}

func removePreviewContainer() {
	docker.RemoveContainer(PreviewContainer)
}

// Preview the website locally using a Docker container.
func DocsPreview() {
	mg.Deps(removePreviewContainer)
	operatorRepo := ensureOperatorRepository()
	operatorDocs, err := filepath.Abs(filepath.Join(operatorRepo, "docs/content"))
	mgx.Must(err)

	// TODO: run on a random port, and then read the output to get the container id and then retrieve the port used

	currentUser, err := user.Current()
	if err != nil {
		currentUser = &user.User{Uid: "0", Gid: "0"}
	}
	setDockerUser := fmt.Sprintf("--user=%s:%s", currentUser.Uid, currentUser.Gid)
	pwd, _ := os.Getwd()
	must.Run("docker", "run", "-d", "-v", pwd+":/src",
		"-v", operatorDocs+":/src/docs/content/operator",
		setDockerUser,
		"-p", "1313:1313", "--name", PreviewContainer, "-w", "/src/docs",
		"klakegg/hugo:0.78.1-ext-alpine", "server", "-D", "-F", "--noHTTPCache",
		"--watch", "--bind=0.0.0.0")

	for {
		output, _ := must.OutputS("docker", "logs", "porter-docs")
		if strings.Contains(output, "Web Server is available") {
			break
		}
		time.Sleep(time.Second)
	}

	must.Run("open", "http://localhost:1313/docs/")
}

// Build a branch preview of the website.
func DocsBranchPreview() {
	setBranchBaseURL()
	Docs()
}

func setBranchBaseURL() {
	// Change the branch to a URL safe slug, e.g. release/v1 -> release-v1
	escapedBranch := strings.ReplaceAll(os.Getenv("BRANCH"), "/", "-")
	// Append a trailing / to the URL for use in Hugo
	baseURL := fmt.Sprintf("https://%s.getporter.org/", escapedBranch)
	os.Setenv("BASEURL", baseURL)
}

// Build a pull request preview of the website.
func DocsPullRequestPreview() {
	setPullRequestBaseURL()
	Docs()
}

func setPullRequestBaseURL() {
	// Append a trailing / to netlify preview domain name for use in Hugo
	baseURL := fmt.Sprintf("%s/", os.Getenv("DEPLOY_PRIME_URL"))
	os.Setenv("BASEURL", baseURL)
}

// clone the other doc repos if they don't exist
// use a local copy as defined in PORTER_OPERATOR_REPOSITORY if available
func linkOperatorDocs() {
	// Remove the old symlink in case the source has moved
	operatorSymlink := "docs/content/operator"
	err := os.RemoveAll(operatorSymlink)
	if !os.IsNotExist(err) {
		mgx.Must(err)
	}

	repoPath := ensureOperatorRepository()
	contentPath, _ := filepath.Abs("docs/content")
	relPath, _ := filepath.Rel(contentPath, filepath.Join(repoPath, "docs/content"))
	log.Println("ln -s", relPath, operatorSymlink)
	mgx.Must(os.Symlink(relPath, operatorSymlink))
}

// Ensures that we have an operator repository and returns its location
func ensureOperatorRepository() string {
	repoPath, err := ensureOperatorRepositoryIn(os.Getenv(LocalOperatorRepositoryEnv), DefaultOperatorSourceDir)
	mgx.Must(err)
	return repoPath
}

// Checks if the repository in localRepo exists and return it
// otherwise clone the repository into defaultRepo, updating with the latest changes if already cloned.
func ensureOperatorRepositoryIn(localRepo string, defaultRepo string) (string, error) {
	// Check if we are using a local repo
	if localRepo != "" {
		if _, err := os.Stat(localRepo); err != nil {
			log.Printf("%s %s does not exist, ignoring\n", LocalOperatorRepositoryEnv, localRepo)
			os.Unsetenv(LocalOperatorRepositoryEnv)
		} else {
			log.Printf("Using operator repository at %s\n", localRepo)
			return localRepo, nil
		}
	}

	// Clone the repo, and ensure it is up-to-date
	cloneDestination, _ := filepath.Abs(defaultRepo)
	_, err := os.Stat(filepath.Join(cloneDestination, ".git"))
	if err != nil && !os.IsNotExist(err) {
		return "", err
	} else if err == nil {
		log.Println("Operator repository already cloned, updating")
		if err = shx.Command("git", "fetch").In(cloneDestination).Run(); err != nil {
			return "", err
		}
		if err = shx.Command("git", "reset", "--hard", "FETCH_HEAD").In(cloneDestination).Run(); err != nil {
			return "", err
		}
		return cloneDestination, nil
	}

	log.Println("Cloning operator repository")
	os.RemoveAll(cloneDestination) // if the path existed but wasn't a git repo, we want to remove it and start fresh
	if err = shx.Run("git", "clone", "https://github.com/getporter/operator.git", cloneDestination); err != nil {
		return "", err
	}
	return cloneDestination, nil
}
