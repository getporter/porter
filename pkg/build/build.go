package build

import (
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

	// INJECT_PORTER_MIXINS_TOKEN can control where mixin instructions will be placed in Dockerfile.
	INJECT_PORTER_MIXINS_TOKEN = "# PORTER_MIXINS"
)

type Builder interface {
	// BuildInvocationImage using the bundle in the build context directory
	BuildInvocationImage(manifest *manifest.Manifest) error

	// TagInvocationImage using the origTag and newTag values supplied
	TagInvocationImage(origTag, newTag string) error
}
