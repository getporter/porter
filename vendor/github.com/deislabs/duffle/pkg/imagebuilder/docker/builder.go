package docker

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/deislabs/duffle/pkg/duffle/manifest"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/image/build"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/builder/dockerignore"
	dockerclient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"

	"github.com/sirupsen/logrus"

	"golang.org/x/net/context"
)

const (
	// DockerignoreFilename is the filename for Docker's ignore file.
	DockerignoreFilename = ".dockerignore"
)

// Builder contains all information to build a docker container image
type Builder struct {
	name         string
	Image        string
	Dockerfile   string
	BuildContext io.ReadCloser

	dockerBuilder dockerBuilder
}

// Builder contains information about the Docker build environment
type dockerBuilder struct {
	DockerClient command.Cli
}

// Name is the name of the image to build
func (db Builder) Name() string {
	return db.name
}

// Type represents the image type to build
func (db Builder) Type() string {
	return "docker"
}

// URI returns the image in the format <registry>/<image>
func (db Builder) URI() string {
	return db.Image
}

// Digest returns the name of a Docker Builder, which will give the image name
//
// TODO - return the actual digest
func (db Builder) Digest() string {
	return strings.Split(db.Image, ":")[1]
}

// NewBuilder returns a new Docker builder based on the manifest
func NewBuilder(c *manifest.InvocationImage, cli *command.DockerCli) *Builder {
	return &Builder{
		name: c.Name,
		// TODO - handle different Dockerfile names
		Dockerfile:    "Dockerfile",
		dockerBuilder: dockerBuilder{DockerClient: cli},
	}
}

// PrepareBuild archives the app directory and loads it as Docker context
func (db *Builder) PrepareBuild(appDir, registry, name string) error {
	if err := archiveSrc(filepath.Join(appDir, db.name), db); err != nil {
		return err
	}

	defer db.BuildContext.Close()

	// write each build context to a buffer so we can also write to the sha256 hash.
	buf := new(bytes.Buffer)
	h := sha256.New()
	w := io.MultiWriter(buf, h)
	if _, err := io.Copy(w, db.BuildContext); err != nil {
		return err
	}

	// truncate checksum to the first 40 characters (20 bytes) this is the
	// equivalent of `shasum build.tar.gz | awk '{print $1}'`.
	ctxtID := h.Sum(nil)
	imgtag := fmt.Sprintf("%.20x", ctxtID)
	imageRepository := path.Join(registry, fmt.Sprintf("%s-%s", name, db.Name()))
	db.Image = fmt.Sprintf("%s:%s", imageRepository, imgtag)

	db.BuildContext = ioutil.NopCloser(buf)

	return nil
}

// Build builds the docker images.
func (db Builder) Build(ctx context.Context, log io.WriteCloser) error {
	defer db.BuildContext.Close()
	buildOpts := types.ImageBuildOptions{
		Tags:       []string{db.Image},
		Dockerfile: db.Dockerfile,
	}

	resp, err := db.dockerBuilder.DockerClient.Client().ImageBuild(ctx, db.BuildContext, buildOpts)
	if err != nil {
		return fmt.Errorf("error building image builder %v with builder %v: %v", db.Name(), db.Type(), err)
	}

	defer resp.Body.Close()
	outFd, isTerm := term.GetFdInfo(db.BuildContext)
	if err := jsonmessage.DisplayJSONMessagesStream(resp.Body, log, outFd, isTerm, nil); err != nil {
		return fmt.Errorf("error streaming messages for image builder %v with builder %v: %v", db.Name(), db.Type(), err)
	}

	if _, _, err = db.dockerBuilder.DockerClient.Client().ImageInspectWithRaw(ctx, db.Image); err != nil {
		if dockerclient.IsErrNotFound(err) {
			return fmt.Errorf("could not locate image for %s: %v", db.Name(), err)
		}
		return fmt.Errorf("imageInspectWithRaw error for image builder %v: %v", db.Name(), err)
	}

	return nil
}

func archiveSrc(contextPath string, b *Builder) error {
	contextDir, relDockerfile, err := build.GetContextFromLocalDir(contextPath, "")
	if err != nil {
		return fmt.Errorf("unable to prepare docker context: %s", err)
	}

	// canonicalize dockerfile name to a platform-independent one
	relDockerfile = archive.CanonicalTarNameForPath(relDockerfile)

	f, err := os.Open(filepath.Join(contextDir, DockerignoreFilename))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	defer f.Close()

	var excludes []string
	if err == nil {
		excludes, err = dockerignore.ReadAll(f)
		if err != nil {
			return err
		}
	}

	if err := build.ValidateContextDirectory(contextDir, excludes); err != nil {
		return fmt.Errorf("error checking docker context: '%s'", err)
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed. The daemon will remove them for us, if needed, after it
	// parses the Dockerfile. Ignore errors here, as they will have been
	// caught by validateContextDirectory above.
	var includes = []string{"."}
	keepThem1, _ := fileutils.Matches(DockerignoreFilename, excludes)
	keepThem2, _ := fileutils.Matches(relDockerfile, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, DockerignoreFilename, relDockerfile)
	}

	logrus.Debugf("INCLUDES: %v", includes)
	logrus.Debugf("EXCLUDES: %v", excludes)
	dockerArchive, err := archive.TarWithOptions(contextDir, &archive.TarOptions{
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	})
	if err != nil {
		return err
	}

	b.name = filepath.Base(contextDir)
	b.BuildContext = dockerArchive
	b.Dockerfile = relDockerfile

	return nil
}
