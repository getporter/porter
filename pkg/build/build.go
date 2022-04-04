package build

import (
	"context"
	"path/filepath"

	"get.porter.sh/porter/pkg/manifest"
)

var (
	// DOCKER_FILE is the file generated before running a docker build.
	DOCKER_FILE = filepath.Join(LOCAL_CNAB, "Dockerfile")

	// LOCAL_CNAB is the generated directory where porter stages the /cnab directory.
	LOCAL_CNAB = ".cnab"

	// LOCAL_APP is the generated directory where porter stages the /cnab/app directory.
	LOCAL_APP = filepath.Join(LOCAL_CNAB, "app")

	// LOCAL_BUNDLE is the generated bundle.json file.
	LOCAL_BUNDLE = filepath.Join(LOCAL_CNAB, "bundle.json")

	// LOCAL_RUN is the path to the generated CNAB entrypoint script, located at /cnab/app/run.
	LOCAL_RUN = filepath.Join(LOCAL_APP, "run")

	// LOCAL_MANIFEST is the canonical Porter manifest generated from the
	// user-provided manifest and any dynamic overrides
	LOCAL_MANIFEST = filepath.Join(LOCAL_APP, "porter.yaml")

	// LOCAL_MIXINS is the path where Porter stages the /cnab/app/mixins directory.
	LOCAL_MIXINS = filepath.Join(LOCAL_APP, "mixins")

	// BUNDLE_DIR is the directory where the bundle is located in the CNAB execution environment.
	BUNDLE_DIR = "/cnab/app"

	// PORTER_MIXINS_TOKEN can control where mixin instructions will be placed in
	// Dockerfile.
	PORTER_MIXINS_TOKEN = "# PORTER_MIXINS"

	// PORTER_INIT_TOKEN controls where Porter's image initialization
	// instructions are placed in the Dockerfile.
	PORTER_INIT_TOKEN = "# PORTER_INIT"
)

type Builder interface {
	// BuildInvocationImage using the bundle in the build context directory
	BuildInvocationImage(ctx context.Context, manifest *manifest.Manifest, opts BuildImageOptions) error

	// TagInvocationImage using the origTag and newTag values supplied
	TagInvocationImage(ctx context.Context, origTag, newTag string) error
}

// BuildImageOptions represents some flags exposed by docker.
type BuildImageOptions struct {
	// SSH is the set of docker build --ssh flags specified.
	SSH []string

	// Secrets is the set of docker build --secret flags specified.
	Secrets []string

	// BuildArgs is the set of docker build --build-arg specified.
	BuildArgs []string

	// NoCache is the docker build --no-cache flag specified.
	NoCache bool
}
