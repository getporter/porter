package docs

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"get.porter.sh/porter/mage/docker"
	"github.com/carolynvs/magex/mgx"
	"github.com/carolynvs/magex/shx"
	"github.com/magefile/mage/mg"
)

var must = shx.CommandBuilder{StopOnError: true}

const (
	LocalOperatorRepositoryEnv = "PORTER_OPERATOR_REPOSITORY"
	PreviewContainer           = "porter-docs"
)

// Generate Porter's static website. Used by Netlify.
// Uses symlinks so it won't work on Windows.
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

// Preview the website documentation.
func DocsPreview() {
	mg.Deps(removePreviewContainer)
	operatorRepo := prepareOperatorRepo()
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

// clone the other doc repos if they don't exist
// use a local copy as defined in PORTER_OPERATOR_REPOSITORY if available
func linkOperatorDocs() {
	docsDest := "docs/content/operator"
	err := os.RemoveAll(docsDest)
	if !os.IsNotExist(err) {
		mgx.Must(err)
	}

	repoPath := prepareOperatorRepo()
	mgx.Must(os.Symlink(filepath.Join(repoPath, "docs/content"), docsDest))
	must.RunV("bash", "-c", "ls "+docsDest)
}

// returns the location of the docs repo
func prepareOperatorRepo() string {
	// Check if we are using a local repo
	if localRepo, ok := os.LookupEnv(LocalOperatorRepositoryEnv); ok {
		if localRepo != "" {
			if _, err := os.Stat(localRepo); err != nil {
				log.Printf("%s %s does not exist, ignoring\n", LocalOperatorRepositoryEnv, localRepo)
				os.Unsetenv(LocalOperatorRepositoryEnv)
			}
		} else {
			log.Printf("Using operator repository at %s\n", localRepo)
			return localRepo
		}
	}

	// Clone the repo
	cloneDestination := "docs/sources/operator"
	_, err := os.Stat(cloneDestination)
	if err == nil { // Already cloned
		log.Println("Operator repository already cloned, skipping")
		return cloneDestination
	}
	if !os.IsNotExist(err) {
		mgx.Must(err)
	}

	log.Println("Cloning operator repository")
	must.Run("git", "clone", "https://github.com/getporter/operator.git", cloneDestination)
	return cloneDestination
}
