package build

import (
	"path/filepath"
)

var (
	// LOCAL_CNAB is the generated directory where porter stages the /cnab directory.
	LOCAL_CNAB = ".cnab"

	// LOCAL_APP is the generated directory where porter stages the /cnab/app directory.
	LOCAL_APP = filepath.Join(LOCAL_CNAB, "app")

	// LOCAL_BUNDLE is the generated bundle.json file.
	LOCAL_BUNDLE = filepath.Join(LOCAL_CNAB, "bundle.json")

	// LOCAL_RUN is the path to the generated CNAB entrypoint script, located at /cnab/app/run.
	LOCAL_RUN = filepath.Join(LOCAL_APP, "run")

	// BUNDLE_DIR is the directory where the bundle is located in the CNAB execution environment.
	BUNDLE_DIR = "/cnab/app"

	// INJECT_PORTER_MIXINS_TOKEN can control where mixin instructions will be placed in Dockerfile.
	INJECT_PORTER_MIXINS_TOKEN = "# PORTER_MIXINS"
)
